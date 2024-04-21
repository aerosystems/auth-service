package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	Mode                    string `mapstructure:"MODE" required:"true"`
	PostgresDSN             string `mapstructure:"POSTGRES_DSN" required:"true"`
	RedisPassword           string `mapstructure:"REDIS_PASSWORD" required:"true"`
	RedisDSN                string `mapstructure:"REDIS_DSN" required:"true"`
	CheckmailServiceRPCAddr string `mapstructure:"CHECKMAIL_SERVICE_RPC_ADDR" required:"true"`
	MailServiceRPCAddr      string `mapstructure:"MAIL_SERVICE_RPC_ADDR" required:"true"`
	CustomerServiceRPCAddr  string `mapstructure:"CUSTOMER_SERVICE_RPC_ADDR" required:"true"`
	AccessSecret            string `mapstructure:"ACCESS_SECRET" required:"true"`
	AccessExpMinutes        int    `mapstructure:"ACCESS_EXP_MINUTES" required:"true"`
	RefreshSecret           string `mapstructure:"REFRESH_SECRET" required:"true"`
	RefreshExpMinutes       int    `mapstructure:"REFRESH_EXP_MINUTES" required:"true"`
	CodeExpMinutes          int    `mapstructure:"CODE_EXP_MINUTES" required:"true"`
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
