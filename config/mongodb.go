package config

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const mongoConnectionTimeout = 10 * time.Second

type MongoDBConnection struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func SetUpMongoDBConnection() (*MongoDBConnection, error) {
	uri := getEnvOrDefault("MONGO_URI", "mongodb://localhost:27017")
	databaseName := getEnvOrDefault("MONGO_DATABASE", "go_gin_clean")

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to MongoDB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoConnectionTimeout)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return &MongoDBConnection{
		Client:   client,
		Database: client.Database(databaseName),
	}, nil
}

func (connection *MongoDBConnection) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoConnectionTimeout)
	defer cancel()

	return connection.Client.Disconnect(ctx)
}
