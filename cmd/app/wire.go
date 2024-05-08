//go:build wireinject
// +build wireinject

package main

import (
	"github.com/aerosystems/auth-service/internal/config"
	rpcRepo "github.com/aerosystems/auth-service/internal/infrastructure/adapters/rpc"
	"github.com/aerosystems/auth-service/internal/infrastructure/repository/pg"
	"github.com/aerosystems/auth-service/internal/models"
	HttpServer "github.com/aerosystems/auth-service/internal/presenters/http"
	"github.com/aerosystems/auth-service/internal/presenters/http/handlers"
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
		wire.Bind(new(handlers.AuthUsecase), new(*usecases.AuthUsecase)),
		wire.Bind(new(handlers.TokenUsecase), new(*usecases.TokenUsecase)),
		wire.Bind(new(usecases.CodeRepository), new(*pg.CodeRepo)),
		wire.Bind(new(usecases.UserRepository), new(*pg.UserRepo)),
		wire.Bind(new(usecases.CheckmailAdapter), new(*rpcRepo.CheckmailAdapter)),
		wire.Bind(new(usecases.MailAdapter), new(*rpcRepo.MailAdapter)),
		wire.Bind(new(usecases.CustomerAdapter), new(*rpcRepo.CustomerAdapter)),
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
		ProvideAuthUsecase,
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

func ProvideUserHandler(baseHandler *handlers.BaseHandler, tokenUsecase handlers.TokenUsecase, authUsecase handlers.AuthUsecase) *handlers.UserHandler {
	panic(wire.Build(handlers.NewUserHandler))
}

func ProvideTokenHandler(baseHandler *handlers.BaseHandler, tokenUsecase handlers.TokenUsecase) *handlers.TokenHandler {
	panic(wire.Build(handlers.NewTokenHandler))
}

func ProvideAuthUsecase(codeRepo usecases.CodeRepository, userRepo usecases.UserRepository, checkmailRepo usecases.CheckmailAdapter, mailRepo usecases.MailAdapter, customerRepo usecases.CustomerAdapter, cfg *config.Config) *usecases.AuthUsecase {
	return usecases.NewAuthUsecase(codeRepo, userRepo, checkmailRepo, mailRepo, customerRepo, cfg.CodeExpMinutes)
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

func ProvideCheckmailRepo(cfg *config.Config) *rpcRepo.CheckmailAdapter {
	rpcClient := RpcClient.NewClient("tcp", cfg.CheckmailServiceRPCAddr)
	return rpcRepo.NewCheckmailAdapter(rpcClient)
}

func ProvideMailRepo(cfg *config.Config) *rpcRepo.MailAdapter {
	rpcClient := RpcClient.NewClient("tcp", cfg.MailServiceRPCAddr)
	return rpcRepo.NewMailAdapter(rpcClient)
}

func ProvideCustomerRepo(cfg *config.Config) *rpcRepo.CustomerAdapter {
	rpcClient := RpcClient.NewClient("tcp", cfg.CustomerServiceRPCAddr)
	return rpcRepo.NewCustomerAdapter(rpcClient)
}
