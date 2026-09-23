package module

import (
	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/internal/handler"
	route "emc_lb/src/internal/routes"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/config"
)

type RBACModule struct {
	routes route.Route
}

func NewRBACModule(cfg *config.AppConfig, queries *sqlc.Queries) *RBACModule {
	rbacService := service.NewRBACService(queries)
	rbacHandler := handler.NewRBACHandler(rbacService)
	rbacRoute := route.NewRBACRoute(rbacHandler, cfg)

	return &RBACModule{
		routes: rbacRoute,
	}
}

func (m *RBACModule) Routes() route.Route {
	return m.routes
}
