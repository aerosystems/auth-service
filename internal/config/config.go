package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	Mode                    string `mapstructure:"MODE"`
	PostgresDSN             string `mapstructure:"POSTGRES_DSN"`
	RedisPassword           string `mapstructure:"REDIS_PASSWORD"`
	RedisDSN                string `mapstructure:"REDIS_DSN"`
	CheckmailServiceRPCAddr string `mapstructure:"CHECKMAIL_SERVICE_RPC_ADDR"`
	MailServiceRPCAddr      string `mapstructure:"MAIL_SERVICE_RPC_ADDR"`
	CustomerServiceRPCAddr  string `mapstructure:"CUSTOMER_SERVICE_RPC_ADDR"`
	AccessSecret            string `mapstructure:"ACCESS_SECRET"`
	AccessExpMinutes        int    `mapstructure:"ACCESS_EXP_MINUTES"`
	RefreshSecret           string `mapstructure:"REFRESH_SECRET"`
	RefreshExpMinutes       int    `mapstructure:"REFRESH_EXP_MINUTES"`
	CodeExpMinutes          int    `mapstructure:"CODE_EXP_MINUTES"`
}

func NewConfig() *Config {
	var cfg Config
	viper.AutomaticEnv()
	executablePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	executableDir := filepath.Dir(executablePath)
	viper.SetConfigFile(filepath.Join(executableDir, ".env"))
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
