package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"e-commerce-go/internal/shared/config"
	"e-commerce-go/pkg/logger"
)

// RedisClient is our wrapper that holds the open Redis connection.
// We do not expose the *redis.Client directly so we can add extra methods if needed.
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient is the constructor function. It reads the configuration and attempts to connect.
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	// 1. Configure Redis connection options (Where is Redis running?)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password, // If Redis has no password, send an empty string
		DB:       cfg.Redis.DB,       // Redis has databases indexed from 0 to 15. Default is 0.

		// Timeout settings (Important to prevent the API from blocking)
		DialTimeout:  5 * time.Second, // Fail if connection takes longer than 5s
		ReadTimeout:  3 * time.Second, // Fail if Redis takes longer than 3s to respond
		WriteTimeout: 3 * time.Second, // Fail if writing takes longer than 3s
	})

	// 2. Ping test (Handshake)
	// Go requires a Context for everything — think of it as a timer.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send a PING. If Redis replies PONG, everything is OK.
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.L().Info("successfully connected to redis",
		zap.String("addr", cfg.Redis.Host),
		zap.Int("db", cfg.Redis.DB),
	)

	return &RedisClient{Client: rdb}, nil
}

// Close gracefully closes the Redis connection when the server shuts down.
func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}
