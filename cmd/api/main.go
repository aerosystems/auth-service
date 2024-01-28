package main

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/middleware"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository"
	RPCServices "github.com/aerosystems/auth-service/internal/rpc_services"
	"github.com/aerosystems/auth-service/internal/services"
	"github.com/aerosystems/auth-service/internal/validators"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	"github.com/aerosystems/auth-service/pkg/logger"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const webPort = 80

// @title Auth Service
// @version 1.0.8
// @description A mandatory part of any microservice infrastructure of a modern WEB application, which is responsible for user authentication and authorization.

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Should contain Access JWT Token, with the Bearer started

// @host gw.verifire.dev/auth
// @schemes https
// @BasePath /
func main() {
	log := logger.NewLogger(os.Getenv("HOSTNAME"))

	clientGORM := GormPostgres.NewClient(logrus.NewEntry(log.Logger))
	_ = clientGORM.AutoMigrate(models.Code{}, models.User{})

	clientREDIS := RedisClient.NewClient()

	checkmailClientRPC := RPCClient.NewClient("tcp", "checkmail-service:5001")
	checkmailRPC := RPCServices.NewCheckmailRPC(checkmailClientRPC)

	mailClientRPC := RPCClient.NewClient("tcp", "mail-service:5001")
	mailRPC := RPCServices.NewMailRPC(mailClientRPC)

	customerClientRPC := RPCClient.NewClient("tcp", "customer-service:5001")
	customerRPC := RPCServices.NewCustomerRPC(customerClientRPC)

	codeRepo := repository.NewCodeRepo(clientGORM)
	userRepo := repository.NewUserRepo(clientGORM)

	userService := services.NewUserServiceImpl(codeRepo, userRepo, checkmailRPC, mailRPC, customerRPC)
	codeService := services.NewCodeServiceImpl(codeRepo)

	accessExpMinutes, _ := strconv.Atoi(os.Getenv("ACCESS_EXP_MINUTES"))
	refreshExpMinutes, _ := strconv.Atoi(os.Getenv("REFRESH_EXP_MINUTES"))
	tokenService := services.NewTokenServiceImpl(clientREDIS, os.Getenv("ACCESS_SECRET"), os.Getenv("REFRESH_SECRET"), accessExpMinutes, refreshExpMinutes)

	baseHandler := handlers.NewBaseHandler(os.Getenv("APP_ENV"), log.Logger, tokenService, userService, codeService)

	oauthMiddleware := middleware.NewOAuthMiddlewareImpl(tokenService)
	basicAuthMiddleware := middleware.NewBasicAuthMiddlewareImpl(os.Getenv("BASIC_AUTH_DOCS_USERNAME"), os.Getenv("BASIC_AUTH_DOCS_PASSWORD"))

	app := NewConfig(baseHandler, oauthMiddleware, basicAuthMiddleware)
	e := app.NewRouter()
	middleware.AddLog(e, log.Logger)

	validator := validator.New()
	validator.RegisterValidation("customPasswordRule", validators.CustomPasswordRule)
	e.Validator = &validators.CustomValidator{Validator: validator}

	log.Infof("starting auth-service HTTP server on port %d\n", webPort)
	if err := e.Start(fmt.Sprintf(":%d", webPort)); err != nil {
		log.Fatal(err)
	}
}
