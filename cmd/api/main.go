package main

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository"
	RPCServices "github.com/aerosystems/auth-service/internal/rpc_services"
	"github.com/aerosystems/auth-service/internal/services"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	"github.com/aerosystems/auth-service/pkg/logger"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

const webPort = "80"

// @title Auth Service
// @version 1.0.7
// @description A mandatory part of any microservice infrastructure of a modern WEB application, which is responsible for user authentication and authorization.

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Should contain Access JWT Token, with the Bearer started

// @host gw.verifire.com/auth
// @schemes https
// @BasePath /
func main() {
	log := logger.NewLogger(os.Getenv("HOSTNAME"))

	clientGORM := GormPostgres.NewClient(logrus.NewEntry(log.Logger))
	_ = clientGORM.AutoMigrate(models.Code{})

	clientREDIS := RedisClient.NewClient()

	checkmailClientRPC := RPCClient.NewClient("tcp", "checkmail-service:5001")
	checkmailRPC := RPCServices.NewCheckmailRPC(checkmailClientRPC)

	mailClientRPC := RPCClient.NewClient("tcp", "mail-service:5001")
	mailRPC := RPCServices.NewMailRPC(mailClientRPC)

	projectClientRPC := RPCClient.NewClient("tcp", "project-service:5001")
	projectRPC := RPCServices.NewProjectRPC(projectClientRPC)

	subscriptionClientRPC := RPCClient.NewClient("tcp", "subscription-service:5001")
	subscriptionRPC := RPCServices.NewSubscriptionRPC(subscriptionClientRPC)

	userClientRPC := RPCClient.NewClient("tcp", "user-service:5001")
	userRPC := RPCServices.NewUserRPC(userClientRPC)

	codeRepo := repository.NewCodeRepo(clientGORM)

	userService := services.NewUserServiceImpl(codeRepo, checkmailRPC, mailRPC, projectRPC, subscriptionRPC, userRPC)
	tokenService := services.NewTokenServiceImpl(clientREDIS)

	app := Config{
		BaseHandler: handlers.NewBaseHandler(
			log.Logger,
			codeRepo,
			tokenService,
			userService,
		),
		TokenService: tokenService,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(log.Logger),
	}

	log.Infof("starting auth-service HTTP Server on port %s\n", webPort)
	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}
