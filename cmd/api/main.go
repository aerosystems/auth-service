package main

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
	"log"
	"net/http"
)

// @title Auth Service
// @version 1.0.5
// @description A mandatory part of any microservice infrastructure of a modern WEB application

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Should contain Access JWT Token, with the Bearer started

// @host localhost:8081
// @BasePath /
func main() {
	clientGORM := GormPostgres.NewClient()
	clientGORM.AutoMigrate(models.User{}, models.Code{})
	clientREDIS := RedisClient.NewClient()

	userRepo := repository.NewUserRepo(clientGORM, clientREDIS)
	codeRepo := repository.NewCodeRepo(clientGORM)

	tokenService := TokenService.NewService(clientREDIS)

	app := Config{
		WebPort: "80",
		BaseHandler: handlers.NewBaseHandler(userRepo,
			codeRepo,
			tokenService,
		),
		TokenService: tokenService,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.WebPort),
		Handler: app.routes(),
	}

	log.Printf("Starting authentication end service on port %s\n", app.WebPort)
	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}
