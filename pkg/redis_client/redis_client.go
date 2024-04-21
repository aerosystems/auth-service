package RedisClient

import (
	"github.com/aerosystems/auth-service/pkg/logger"
	"github.com/go-redis/redis/v7"
	"time"
)

func NewRedisClient(log *logger.Logger, dsn, password string) *redis.Client {
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
