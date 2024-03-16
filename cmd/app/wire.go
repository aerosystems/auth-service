//go:build wireinject
// +build wireinject

package main

import (
	"github.com/aerosystems/auth-service/internal/config"
	HttpServer "github.com/aerosystems/auth-service/internal/infrastructure/http"
	"github.com/aerosystems/auth-service/internal/infrastructure/http/handlers"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository/pg"
	rpcRepo "github.com/aerosystems/auth-service/internal/repository/rpc"
	"github.com/aerosystems/auth-service/internal/usecases"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	"github.com/aerosystems/auth-service/pkg/logger"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	RpcClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/go-redis/redis/v7"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//go:generate wire
func InitApp() *App {
	panic(wire.Build(
		wire.Bind(new(handlers.UserUsecase), new(*usecases.UserUsecase)),
		wire.Bind(new(handlers.CodeUsecase), new(*usecases.CodeUsecase)),
		wire.Bind(new(handlers.TokenUsecase), new(*usecases.TokenUsecase)),
		wire.Bind(new(usecases.CodeRepository), new(*pg.CodeRepo)),
		wire.Bind(new(usecases.UserRepository), new(*pg.UserRepo)),
		wire.Bind(new(usecases.CheckmailRepo), new(*rpcRepo.CheckmailRepo)),
		wire.Bind(new(usecases.MailRepo), new(*rpcRepo.MailRepo)),
		wire.Bind(new(usecases.CustomerRepo), new(*rpcRepo.CustomerRepo)),
		ProvideApp,
		ProvideLogger,
		ProvideConfig,
		ProvideHttpServer,
		ProvideLogrusLogger,
		ProvideLogrusEntry,
		ProvideGormPostgres,
		ProvideRedisClient,
		ProvideBaseHandler,
		ProvideUserHandler,
		ProvideTokenHandler,
		ProvideUserUsecase,
		ProvideCodeUsecase,
		ProvideTokenUsecase,
		ProvideCodeRepo,
		ProvideUserRepo,
		ProvideCheckmailRepo,
		ProvideMailRepo,
		ProvideCustomerRepo,
	))
}

func ProvideApp(log *logrus.Logger, cfg *config.Config, httpServer *HttpServer.Server) *App {
	panic(wire.Build(NewApp))
}

func ProvideLogger() *logger.Logger {
	panic(wire.Build(logger.NewLogger))
}

func ProvideConfig() *config.Config {
	panic(wire.Build(config.NewConfig))
}

func ProvideHttpServer(log *logrus.Logger, cfg *config.Config, userHandler *handlers.UserHandler, tokenHandler *handlers.TokenHandler) *HttpServer.Server {
	return HttpServer.NewServer(log, cfg.AccessSecret, userHandler, tokenHandler)
}

func ProvideLogrusEntry(log *logger.Logger) *logrus.Entry {
	return logrus.NewEntry(log.Logger)
}

func ProvideLogrusLogger(log *logger.Logger) *logrus.Logger {
	return log.Logger
}

func ProvideGormPostgres(e *logrus.Entry, cfg *config.Config) *gorm.DB {
	db := GormPostgres.NewClient(e, cfg.PostgresDSN)
	if err := db.AutoMigrate(&models.User{}, &models.Code{}); err != nil { // TODO: Move to migration
		panic(err)
	}
	return db
}

func ProvideRedisClient(log *logger.Logger, cfg *config.Config) *redis.Client {
	return RedisClient.NewRedisClient(log, cfg.RedisDSN, cfg.RedisPassword)
}

func ProvideBaseHandler(log *logrus.Logger, cfg *config.Config) *handlers.BaseHandler {
	return handlers.NewBaseHandler(log, cfg.Mode)
}

func ProvideUserHandler(baseHandler *handlers.BaseHandler, tokenUsecase handlers.TokenUsecase, userUsecase handlers.UserUsecase, codeUsecase handlers.CodeUsecase) *handlers.UserHandler {
	panic(wire.Build(handlers.NewUserHandler))
}

func ProvideTokenHandler(baseHandler *handlers.BaseHandler, tokenUsecase handlers.TokenUsecase) *handlers.TokenHandler {
	panic(wire.Build(handlers.NewTokenHandler))
}

func ProvideUserUsecase(codeRepo usecases.CodeRepository, userRepo usecases.UserRepository, checkmailRepo usecases.CheckmailRepo, mailRepo usecases.MailRepo, customerRepo usecases.CustomerRepo) *usecases.UserUsecase {
	panic(wire.Build(usecases.NewUserUsecase))
}

func ProvideCodeUsecase(codeRepo usecases.CodeRepository) *usecases.CodeUsecase {
	panic(wire.Build(usecases.NewCodeUsecase))
}

func ProvideTokenUsecase(redisClient *redis.Client, cfg *config.Config) *usecases.TokenUsecase {
	return usecases.NewTokenUsecase(redisClient, cfg.AccessSecret, cfg.RefreshSecret, cfg.AccessExpMinutes, cfg.RefreshExpMinutes)
}

func ProvideCodeRepo(db *gorm.DB, cfg *config.Config) *pg.CodeRepo {
	return pg.NewCodeRepo(db, cfg.CodeExpMinutes)
}

func ProvideUserRepo(db *gorm.DB) *pg.UserRepo {
	panic(wire.Build(pg.NewUserRepo))
}

func ProvideCheckmailRepo(cfg *config.Config) *rpcRepo.CheckmailRepo {
	rpcClient := RpcClient.NewClient("tcp", cfg.CheckmailServiceRPCAddr)
	return rpcRepo.NewCheckmailRepo(rpcClient)
}

func ProvideMailRepo(cfg *config.Config) *rpcRepo.MailRepo {
	rpcClient := RpcClient.NewClient("tcp", cfg.MailServiceRPCAddr)
	return rpcRepo.NewMailRepo(rpcClient)
}

func ProvideCustomerRepo(cfg *config.Config) *rpcRepo.CustomerRepo {
	rpcClient := RpcClient.NewClient("tcp", cfg.CustomerServiceRPCAddr)
	return rpcRepo.NewCustomerRepo(rpcClient)
}
