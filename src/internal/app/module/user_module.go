package module

import (
	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/repository"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/storage"
	"emc_lb/src/pkg/worker"
)

type UserModule struct {
	routes route.Route
}

func NewUserModule(cfg *config.AppConfig, queries sqlc.Querier, refreshTokenStore cache.RefreshTokenStore, emailOTPStore cache.EmailOTPStore, taskDistributor worker.TaskDistributor, avatarStorage storage.AvatarStorage) *UserModule {
	userRepository := repository.NewUserRepository(queries)
	userService := service.NewUserService(cfg, userRepository, refreshTokenStore, emailOTPStore, taskDistributor, avatarStorage)
	userHandler := handler.NewUserHandler(userService)
	userRoute := route.NewUserRoute(userHandler)

	return &UserModule{
		routes: userRoute,
	}
}

func (m *UserModule) Routes() route.Route {
	return m.routes
}
