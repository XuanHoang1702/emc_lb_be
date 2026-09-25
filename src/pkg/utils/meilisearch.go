package utils

import (
	"context"
	"fmt"
	"time"

	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/logs"
	"github.com/meilisearch/meilisearch-go"
)

func NewMeilisearchClientFromConfig(cfg *config.SearchSettings) (meilisearch.ServiceManager, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("MEILISEARCH_HOST is missing")
	}
	
	client := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.MasterKey))

	// Health check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthy := false
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("meilisearch health check failed: timeout")
		default:
			if client.IsHealthy() {
				healthy = true
				break
			}
			time.Sleep(1 * time.Second)
		}
		if healthy {
			break
		}
	}

	logs.L().Info("Connected to Meilisearch at " + cfg.Host)
	return client, nil
}
