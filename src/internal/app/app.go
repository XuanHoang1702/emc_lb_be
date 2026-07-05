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
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mail"
	"emc_lb/src/pkg/storage"
	"emc_lb/src/pkg/utils"
	"emc_lb/src/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type App struct {
	server            *http.Server
	pgPool            *pgxpool.Pool
	mongoClient       *mongo.Client
	redisClient       *redis.Client
	queries           *sqlc.Queries
	refreshTokenStore cache.RefreshTokenStore
	avatarStorage     storage.AvatarStorage
}

func New() (*App, error) {
	if err := validation.InitValidator(); err != nil {
		return nil, fmt.Errorf("initialize validator: %w", err)
	}

	databaseURL := utils.BuildDatabaseURL()
	pgPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	mongoClient, err := utils.NewMongoClient(context.Background())
	if err != nil {
		pgPool.Close()
		return nil, fmt.Errorf("create mongo client: %w", err)
	}

	redisClient, err := utils.NewRedisClient()
	if err != nil {
		pgPool.Close()
		_ = mongoClient.Disconnect(context.Background())
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	avatarStorage, err := storage.NewLocalStackS3Storage(context.Background())
	if err != nil {
		pgPool.Close()
		_ = mongoClient.Disconnect(context.Background())
		_ = redisClient.Close()
		return nil, fmt.Errorf("create avatar storage: %w", err)
	}

	if err := avatarStorage.EnsureBucket(context.Background()); err != nil {
		pgPool.Close()
		_ = mongoClient.Disconnect(context.Background())
		_ = redisClient.Close()
		return nil, fmt.Errorf("ensure avatar bucket: %w", err)
	}

	router := gin.New()
	router.Use(middleware.RequestLogMiddleware(), middleware.AcceptLanguageMiddleware(), gin.Recovery())

	// MODULES
	queries := sqlc.New(logs.WrapDBTX(pgPool))
	auth.InitRBACManager(queries)
	refreshTokenStore := cache.NewRedisRefreshTokenStore(redisClient)
	emailOTPStore := cache.NewRedisEmailOTPStore(redisClient)
	mailer := mail.NewSMTPMailer()
	userModule := module.NewUserModule(queries, refreshTokenStore, emailOTPStore, mailer, avatarStorage)
	productModule := module.NewProductModule(mongoClient.Database(utils.GetMongoDatabaseName()))
	categoryModule, err := module.NewCategoryModule(mongoClient.Database(utils.GetMongoDatabaseName()))
	//=================================================================================================

	if err != nil {
		pgPool.Close()
		_ = mongoClient.Disconnect(context.Background())
		_ = redisClient.Close()
		return nil, fmt.Errorf("create category module: %w", err)
	}
	brandModule, err := module.NewBrandModule(mongoClient.Database(utils.GetMongoDatabaseName()))
	if err != nil {
		pgPool.Close()
		_ = mongoClient.Disconnect(context.Background())
		_ = redisClient.Close()
		return nil, fmt.Errorf("create brand module: %w", err)
	}

	orderModule := module.NewOrderModule(mongoClient.Database(utils.GetMongoDatabaseName()))
	paymentModule := module.NewPaymentModule(orderModule.Service())

	route.RegisterRoutes(router, []route.Route{
		userModule.Routes(),
		productModule.Routes(),
		categoryModule.Routes(),
		brandModule.Routes(),
		orderModule.Routes(),
		paymentModule.Routes(),
	})

	server := &http.Server{
		Addr:              ":" + utils.GetEnv("APP_PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

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
	if a.pgPool != nil {
		a.pgPool.Close()
	}

	if a.mongoClient != nil {
		_ = a.mongoClient.Disconnect(ctx)
	}

	if a.redisClient != nil {
		_ = a.redisClient.Close()
	}

	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}
