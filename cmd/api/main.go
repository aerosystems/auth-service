package main

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository"
	"github.com/aerosystems/auth-service/pkg/mygorm"
	"github.com/aerosystems/auth-service/pkg/myredis"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct {
	BaseHandler *handlers.BaseHandler
	TokensRepo  models.TokensRepository
}

// @title Auth Service
// @version 1.0
// @description A mandatory part of any microservice infrastructure of a modern WEB application

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /v1
func main() {
	clientGORM := mygorm.NewClient()
	clientREDIS := myredis.NewClient()
	userRepo := repository.NewUserRepo(clientGORM, clientREDIS)
	codeRepo := repository.NewCodeRepo(clientGORM)
	tokensRepo := repository.NewTokensRepo(clientREDIS)

	app := Config{
		BaseHandler: handlers.NewBaseHandler(userRepo,
			codeRepo,
			tokensRepo,
		),
		TokensRepo: tokensRepo,
	}

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
