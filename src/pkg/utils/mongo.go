package utils

import (
	"context"
	"fmt"
	"time"

	"emc_lb/src/pkg/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewMongoClientFromConfig creates a MongoDB client using typed AppConfig.
func NewMongoClientFromConfig(ctx context.Context, cfg *config.MongoSettings) (*mongo.Client, error) {
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute).
		SetMaxConnecting(10).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Database(cfg.Database).RunCommand(pingCtx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return nil, fmt.Errorf("ping mongo %s: %w", cfg.Database, err)
	}

	return client, nil
}

// NewMongoClient creates a MongoDB client using env vars directly (legacy fallback).
// Prefer NewMongoClientFromConfig when AppConfig is available.
func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	uri, err := BuildMongoURI()
	if err != nil {
		return nil, err
	}
	cfg := &config.MongoSettings{
		URI:      uri,
		Database: GetMongoDatabaseName(),
	}
	return NewMongoClientFromConfig(ctx, cfg)
}

// BuildMongoURI returns the MongoDB connection URI from environment variables.
// It checks MONGO_URI first, then falls back to MONGO_INITDB_ROOT_URL.
func BuildMongoURI() (string, error) {
	if value := GetEnv("MONGO_URI", ""); value != "" {
		return value, nil
	}

	if value := GetEnv("MONGO_INITDB_ROOT_URL", ""); value != "" {
		return value, nil
	}

	return "", fmt.Errorf("mongo: MONGO_URI or MONGO_INITDB_ROOT_URL environment variable is required")
}

// GetMongoDatabaseName returns the MongoDB database name from the MONGO_DB
// environment variable, defaulting to "emc_lb".
func GetMongoDatabaseName() string {
	if value := GetEnv("MONGO_DB", ""); value != "" {
		return value
	}

	return "emc_lb"
}
