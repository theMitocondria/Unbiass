package inits

import (
	"os"
	"log"
	"github.com/redis/go-redis/v9"
	"context"
)

var RedisClient *redis.Client

func Redis(){
	RedisClient = redis.NewClient(&redis.Options{
		Addr : os.Getenv("REDIS_HOST"),
		Password : "",
		DB : 0,
	})

	_,err:= RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis %v",err)
	}
}