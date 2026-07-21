package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var CTX = context.Background()
var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := RDB.Ping(CTX).Result()
	if err != nil {
		panic(err)
	}
}
