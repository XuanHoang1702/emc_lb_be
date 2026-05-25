package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"emc_lb/src/internal/app"
	"emc_lb/src/pkg/utils"
)

func main() {
	utils.LoadEnv()

	application, err := app.New()
	if err != nil {
		log.Fatalf("unable to initialize app: %v", err)
	}

	go func() {
		stopSignal := make(chan os.Signal, 1)
		signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)
		<-stopSignal

		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if shutdownErr := application.Close(shutdownContext); shutdownErr != nil {
			log.Printf("graceful shutdown failed: %v", shutdownErr)
		}
	}()

	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped unexpectedly: %v", err)
	}
}
