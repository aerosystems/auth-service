//go:build wireinject
// +build wireinject

package main

import (
	"github.com/aerosystems/auth-service/internal/config"
	HTTPServer "github.com/aerosystems/auth-service/internal/http"
	"github.com/aerosystems/auth-service/internal/infrastructure/rest"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/repository/pg"
	rpcRepo "github.com/aerosystems/auth-service/internal/repository/rpc"
	"github.com/aerosystems/auth-service/internal/usecases"
	GormPostgres "github.com/aerosystems/auth-service/pkg/gorm_postgres"
	"github.com/aerosystems/auth-service/pkg/logger"
	OAuthService "github.com/aerosystems/auth-service/pkg/oauth"
	RedisClient "github.com/aerosystems/auth-service/pkg/redis_client"
	RPCClient "github.com/aerosystems/auth-service/pkg/rpc_client"
	"github.com/go-redis/redis/v7"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//go:generate wire
func InitApp() *App {
	panic(wire.Build(
		wire.Bind(new(rest.UserUsecase), new(*usecases.UserUsecase)),
		wire.Bind(new(rest.CodeUsecase), new(*usecases.CodeUsecase)),
		wire.Bind(new(rest.TokenUsecase), new(*usecases.TokenUsecase)),
		wire.Bind(new(usecases.CodeRepository), new(*pg.CodeRepo)),
		wire.Bind(new(usecases.UserRepository), new(*pg.UserRepo)),
		wire.Bind(new(usecases.CheckmailRepo), new(*rpcRepo.CheckmailRepo)),
		wire.Bind(new(usecases.MailRepo), new(*rpcRepo.MailRepo)),
		wire.Bind(new(usecases.CustomerRepo), new(*rpcRepo.CustomerRepo)),
		wire.Bind(new(HTTPServer.TokenService), new(*OAuthService.AccessTokenService)),
		ProvideApp,
		ProvideLogger,
		ProvideConfig,
		ProvideHTTPServer,
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
		ProvideAccessTokenService,
	))
}

func ProvideApp(log *logrus.Logger, cfg *config.Config, httpServer *HTTPServer.Server) *App {
	panic(wire.Build(NewApp))
}

func ProvideLogger() *logger.Logger {
	panic(wire.Build(logger.NewLogger))
}

func ProvideConfig() *config.Config {
	panic(wire.Build(config.NewConfig))
}

func ProvideHTTPServer(log *logrus.Logger, userHandler *rest.UserHandler, tokenHandler *rest.TokenHandler, tokenService HTTPServer.TokenService) *HTTPServer.Server {
	panic(wire.Build(HTTPServer.NewServer))
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

func ProvideBaseHandler(log *logrus.Logger, cfg *config.Config) *rest.BaseHandler {
	return rest.NewBaseHandler(log, cfg.Mode)
}

func ProvideUserHandler(baseHandler *rest.BaseHandler, tokenUsecase rest.TokenUsecase, userUsecase rest.UserUsecase, codeUsecase rest.CodeUsecase) *rest.UserHandler {
	panic(wire.Build(rest.NewUserHandler))
}

func ProvideTokenHandler(baseHandler *rest.BaseHandler, tokenUsecase rest.TokenUsecase) *rest.TokenHandler {
	panic(wire.Build(rest.NewTokenHandler))
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
	rpcClient := RPCClient.NewClient("tcp", cfg.CheckmailServiceRPCAddr)
	return rpcRepo.NewCheckmailRepo(rpcClient)
}

func ProvideMailRepo(cfg *config.Config) *rpcRepo.MailRepo {
	rpcClient := RPCClient.NewClient("tcp", cfg.MailServiceRPCAddr)
	return rpcRepo.NewMailRepo(rpcClient)
}

func ProvideCustomerRepo(cfg *config.Config) *rpcRepo.CustomerRepo {
	rpcClient := RPCClient.NewClient("tcp", cfg.CustomerServiceRPCAddr)
	return rpcRepo.NewCustomerRepo(rpcClient)
}

func ProvideAccessTokenService(cfg *config.Config) *OAuthService.AccessTokenService {
	return OAuthService.NewAccessTokenService(cfg.AccessSecret)
}
