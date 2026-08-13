package config

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisConnectionTimeout = 10 * time.Second

type RedisConnection struct {
	Client *redis.Client
}

func SetUpRedisConnection() (*RedisConnection, error) {
	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	username := getEnvOrDefault("REDIS_USERNAME", "")
	password := getEnvOrDefault("REDIS_PASSWORD", "")
	dbValue := getEnvOrDefault("REDIS_DB", "0")

	db, err := strconv.Atoi(dbValue)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_DB: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisConnectionTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return &RedisConnection{Client: client}, nil
}

func (connection *RedisConnection) Shutdown() error {
	return connection.Client.Close()
}
