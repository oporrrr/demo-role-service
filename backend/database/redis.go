package database

import (
	"context"
	"log"

	"demo-role-service/config"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() {
	opt, err := redis.ParseURL(config.Cfg.RedisURL)
	if err != nil {
		log.Printf("redis: invalid URL, cache disabled: %v", err)
		return
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis: connection failed, cache disabled: %v", err)
		return
	}
	Redis = client
	log.Println("redis: connected")
}
