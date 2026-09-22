package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mail"
	"emc_lb/src/pkg/utils"
	"emc_lb/src/pkg/worker"

	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	logs.Init(cfg.App.LogLevel, cfg.App.Mode)
	logger := logs.L()
	logger.Info("Starting worker server...")

	pgPool, err := utils.NewPgPoolFromConfig(context.Background(), &cfg.Postgres)
	if err != nil {
		logger.Error("failed to create postgres pool", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	_ = sqlc.New(logs.WrapDBTX(pgPool))

	mailer := mail.NewMailer()

	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
	}

	processor := worker.NewRedisTaskProcessor(redisOpt, mailer)

	err = processor.Start()
	if err != nil {
		logger.Error("failed to start worker server", "error", err)
		os.Exit(1)
	}
	logger.Info("worker server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("shutting down worker server", "signal", sig.String())

	processor.Shutdown()
	logger.Info("worker server stopped")
}
