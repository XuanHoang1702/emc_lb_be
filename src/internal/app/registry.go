package app

import (
	"emc_lb/src/internal/db/sqlc"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/mail"
	"emc_lb/src/pkg/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AppDeps is a dependency container holding all shared resources.
// It is constructed once in app.New() and passed to each ModuleFactory.
// Modules should only read from AppDeps, never mutate it.
type AppDeps struct {
	Config *config.AppConfig

	// Postgres
	PgPool    *pgxpool.Pool
	PgQueries *sqlc.Queries

	// MongoDB
	MongoClient *mongo.Client
	MongoDB     *mongo.Database

	// Redis
	RedisClient *redis.Client

	// Cache stores (built on Redis)
	RefreshTokenStore cache.RefreshTokenStore
	EmailOTPStore     cache.EmailOTPStore

	// External services
	AvatarStorage storage.AvatarStorage
	Mailer        mail.Mailer
}

// ModuleFactory is a function that constructs a module's route.Route from
// shared dependencies. Each module registers itself via Registry.Register.
type ModuleFactory func(deps *AppDeps) route.Route

// Registry holds the ordered list of module factories.
// Modules are built and registered in the order they were added.
type Registry struct {
	factories []ModuleFactory
}

// NewRegistry creates an empty module registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a module factory to the registry and returns the registry
// for method chaining.
//
// Example:
//
//	registry.
//	    Register(userModule).
//	    Register(productModule).
//	    Register(orderModule)
func (r *Registry) Register(f ModuleFactory) *Registry {
	r.factories = append(r.factories, f)
	return r
}

// Build instantiates all registered modules using the provided deps and
// returns the list of routes ready to be passed to RegisterRoutes.
func (r *Registry) Build(deps *AppDeps) []route.Route {
	routes := make([]route.Route, 0, len(r.factories))

	for _, factory := range r.factories {
		if rt := factory(deps); rt != nil {
			routes = append(routes, rt)
		}
	}

	return routes
}
