package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aerosystems/auth-service/data"
	"github.com/go-redis/redis/v7"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Cache  *redis.Client
	Models data.Models
	//Etcd   *clientv3.Client
}

func main() {
	log.Println("---------------------------------------------")
	log.Println("Attempting to connect to Postgres...")
	// connect to the database
	connDB := connectToDB()
	if connDB == nil {
		log.Panic("can't connect to postgres!")
	}

	log.Println("---------------------------------------------")
	log.Println("Attempting to connect to Redis...")
	// connect to the database
	connCache := connectToCache()
	if connCache == nil {
		log.Panic("can't connect to redis!")
	}

	app := Config{
		DB:     connDB,
		Cache:  connCache,
		Models: data.New(connDB),
	}

	//app.registerService()
	//defer app.Etcd.Close()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Printf("Starting authentication end service on port %s\n", webPort)
	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}

func connectToDB() *sql.DB {
	// connect to postgres
	dsn := os.Getenv("POSTGRES_DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not ready...")
			counts++
		} else {
			log.Println("Connected to database!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}

func connectToCache() *redis.Client {
	dsn := os.Getenv("REDIS_DSN")
	password := os.Getenv("REDIS_PASSWORD")

	for {
		client := redis.NewClient(&redis.Options{
			Addr:     dsn,
			Password: password,
		})

		_, err := client.Ping().Result()

		if err != nil {
			log.Println("Redis not ready....")
			counts++
		} else {
			return client
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
