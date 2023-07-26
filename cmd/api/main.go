package main

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	"github.com/aerosystems/auth-service/pkg/logger"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
	"net/http"
	"net/rpc"
	"os"
)

const webPort = "80"

// @title Auth Service
// @version 1.0.7
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
	log := logger.NewLogger(os.Getenv("HOSTNAME"))

	clientGORM := GormPostgres.NewClient()
	clientGORM.AutoMigrate(models.User{}, models.Code{})
	clientREDIS := RedisClient.NewClient()

	userRepo := repository.NewUserRepo(clientGORM, clientREDIS)
	codeRepo := repository.NewCodeRepo(clientGORM)

	tokenService := TokenService.NewService(clientREDIS)

	projectClientRPC, err := rpc.Dial("tcp", "project-service:5001")
	if err != nil {
		log.Fatal(err)
	}

	mailClientRPC, err := rpc.Dial("tcp", "mail-service:5001")
	if err != nil {
		log.Fatal(err)
	}

	app := Config{
		BaseHandler: handlers.NewBaseHandler(userRepo,
			codeRepo,
			tokenService,
			projectClientRPC,
			mailClientRPC,
		),
		TokenService: tokenService,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Info("starting auth-service HTTP Server on port %s\n", webPort)
	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}
