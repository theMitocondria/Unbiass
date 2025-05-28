package inits

import (
	"os"
	"log"
	"fmt"
	"github.com/redis/go-redis/v9"
	"context"
	"crypto/tls"

)

var RedisClient *redis.Client

func Redis(){

	fmt.Println(os.Getenv("REDIS_HOST"))

	RedisClient = redis.NewClient(&redis.Options{
        Addr:      os.Getenv("REDIS_HOST"), // only host:port
        Username:  "default",                                            // from Aiven
        Password:  os.Getenv("REDIS_PASSWORD"),                             // from Aiven
        TLSConfig: &tls.Config{},                                       // use TLS for `rediss://`
    })

	_,err:= RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis %v",err)
	}
}