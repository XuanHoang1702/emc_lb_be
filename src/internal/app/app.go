package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"emc_lb/src/internal/app/module"
	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/internal/middleware"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/pkg/auth"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mail"
	"emc_lb/src/pkg/migrate"
	"emc_lb/src/pkg/storage"
	"emc_lb/src/pkg/utils"
	"emc_lb/src/pkg/validation"
	"emc_lb/src/pkg/worker"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// App holds all long-lived server resources and exposes Run / Close.
type App struct {
	server            *http.Server
	pgPool            *pgxpool.Pool
	mongoClient       *mongo.Client
	redisClient       *redis.Client
	queries           *sqlc.Queries
	refreshTokenStore cache.RefreshTokenStore
	avatarStorage     storage.AvatarStorage
}

// New bootstraps the full application.
// Order: config → logger → validator → migrate → connections → modules → server.
func New() (*App, error) {
	// ── 1. Load & validate typed config ──────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// ── 2. Structured logger ─────────────────────────────────────────────────
	logs.Init(cfg.App.LogLevel, cfg.App.Mode)

	// ── 3. Input validation ──────────────────────────────────────────────────
	if err = validation.InitValidator(); err != nil {
		return nil, fmt.Errorf("initialize validator: %w", err)
	}

	// Cleanup stack: on failure, close all initialized resources in reverse order.
	var closers []func()
	cleanup := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}
	success := false
	defer func() {
		if !success {
			cleanup()
		}
	}()

	// ── 4. Auto-migration ────────────────────────────────────────────────────
	if err = migrate.Run(cfg.Postgres.DatabaseURL(), cfg.App.MigrationsDir); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// ── 5. Database connections ──────────────────────────────────────────────
	pgPool, err := utils.NewPgPoolFromConfig(context.Background(), &cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	closers = append(closers, func() { pgPool.Close() })

	mongoClient, err := utils.NewMongoClientFromConfig(context.Background(), &cfg.MongoDB)
	if err != nil {
		return nil, fmt.Errorf("create mongo client: %w", err)
	}
	closers = append(closers, func() { _ = mongoClient.Disconnect(context.Background()) })

	redisClient, err := utils.NewRedisClientFromConfig(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}
	closers = append(closers, func() { _ = redisClient.Close() })

	avatarStorage, err := storage.NewLocalStackS3Storage(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create avatar storage: %w", err)
	}

	if err = avatarStorage.EnsureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure avatar bucket: %w", err)
	}

	// ── 6. Shared services ───────────────────────────────────────────────────
	queries := sqlc.New(logs.WrapDBTX(pgPool))
	auth.InitRBACManager(queries)
	refreshTokenStore := cache.NewRedisRefreshTokenStore(redisClient)
	emailOTPStore := cache.NewRedisEmailOTPStore(redisClient)
	mailer := mail.NewSMTPMailer()
	mongoDB := mongoClient.Database(cfg.MongoDB.Database)

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	// ── 7. AppDeps container ─────────────────────────────────────────────────
	deps := &AppDeps{
		Config:            cfg,
		PgPool:            pgPool,
		PgQueries:         queries,
		MongoClient:       mongoClient,
		MongoDB:           mongoDB,
		RedisClient:       redisClient,
		RefreshTokenStore: refreshTokenStore,
		EmailOTPStore:     emailOTPStore,
		AvatarStorage:     avatarStorage,
		Mailer:            mailer,
		TaskDistributor:   taskDistributor,
	}

	// ── 8. Module registry ───────────────────────────────────────────────────
	// Modules are built here. Cross-module dependencies (e.g. couponSvc inside
	// orderModule) are wired via closures that capture module instances.
	// Each factory receives deps and returns a route.Route.
	var (
		couponMod  *module.CouponModule
		productMod *module.ProductModule
	)

	couponMod = module.NewCouponModule(mongoDB)
	productMod = module.NewProductModule(mongoDB, redisClient)

	categoryMod, err := module.NewCategoryModule(mongoDB, redisClient)
	if err != nil {
		return nil, fmt.Errorf("create category module: %w", err)
	}
	brandMod, err := module.NewBrandModule(mongoDB)
	if err != nil {
		return nil, fmt.Errorf("create brand module: %w", err)
	}

	orderMod := module.NewOrderModule(mongoDB, mongoClient, couponMod.ServiceInstance(), redisClient, productMod.CacheStore())
	paymentMod := module.NewPaymentModule(orderMod.Service())
	cartMod := module.NewCartModule(mongoDB, productMod.Repository(), couponMod.ServiceInstance())
	shopMod := module.NewShopModule(mongoDB, redisClient)
	userMod := module.NewUserModule(cfg, queries, refreshTokenStore, emailOTPStore, taskDistributor, avatarStorage)

	_ = deps // deps available for future module factories via registry

	// ── 9. HTTP router ───────────────────────────────────────────────────────
	router := gin.New()
	router.Use(
		middleware.RequestIDMiddleware(),
		middleware.CORSMiddleware(),
		middleware.RequestLogMiddleware(),
		middleware.AcceptLanguageMiddleware(),
		gin.Recovery(),
	)

	route.RegisterRoutes(router, []route.Route{
		userMod.Routes(),
		shopMod.Routes(),
		productMod.Routes(),
		categoryMod.Routes(),
		brandMod.Routes(),
		orderMod.Routes(),
		paymentMod.Routes(),
		cartMod.Route,
		couponMod.Routes(),
	}, redisClient, pgPool, mongoClient, cfg)

	// ── 10. HTTP server ──────────────────────────────────────────────────────
	server := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	success = true
	return &App{
		server:            server,
		pgPool:            pgPool,
		mongoClient:       mongoClient,
		redisClient:       redisClient,
		queries:           queries,
		refreshTokenStore: refreshTokenStore,
		avatarStorage:     avatarStorage,
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Close(ctx context.Context) error {
	var shutdownErr error

	if a.server != nil {
		shutdownErr = a.server.Shutdown(ctx)
	}

	if a.pgPool != nil {
		a.pgPool.Close()
	}

	if a.mongoClient != nil {
		_ = a.mongoClient.Disconnect(ctx)
	}

	if a.redisClient != nil {
		_ = a.redisClient.Close()
	}

	return shutdownErr
}
