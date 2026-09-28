package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"super-app-chonburi-go/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func ConnectRedis(cfg *config.Config) *redis.Client {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to connect to Redis at %s: %v. Caching layer will fallback.", addr, err)
	} else {
		log.Printf("✅ Connected to Redis successfully at %s", addr)
	}

	Client = rdb
	return rdb
}
