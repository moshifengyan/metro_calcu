package redis_service

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "",               // 如果有密码，可以在这里设置
		DB:       0,                // 使用的数据库
	})
}

func set(ctx context.Context, key string, val interface{}) error {
	return redisClient.Set(ctx, key, val, 3600).Err()
}

func get(ctx context.Context, key string) interface{} {
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	return val
}
