package utils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func BuildMongoURI() string {
	if value := GetEnv("MONGO_URI", ""); value != "" {
		return value
	}

	if value := GetEnv("MONGO_INITDB_ROOT_URL", ""); value != "" {
		return value
	}

	return "mongodb://admin:admin@localhost:27017/?authSource=admin"
}

func GetMongoDatabaseName() string {
	if value := GetEnv("MONGO_DB", ""); value != "" {
		return value
	}

	return "emc_lb"
}

func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(BuildMongoURI()))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Database(GetMongoDatabaseName()).RunCommand(pingCtx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return nil, err
	}

	return client, nil
}
