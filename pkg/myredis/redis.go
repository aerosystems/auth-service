package myredis

import (
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v7"
)

func NewClient() *redis.Client {
	dsn := os.Getenv("REDIS_DSN")
	password := os.Getenv("REDIS_PASSWORD")

	count := 0

	for {
		client := redis.NewClient(&redis.Options{
			Addr:     dsn,
			Password: password,
		})

		_, err := client.Ping().Result()

		if err != nil {
			log.Println("Redis not ready....")
			count++
		} else {
			return client
		}

		if count > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}

}
