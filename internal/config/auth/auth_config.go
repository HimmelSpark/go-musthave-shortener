package auth

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	SecretKey *string `env:"AUTH_SECRET_KEY,required"`
}

func Init() *Config {
	cmdSecretKey := flag.String("k", "dev-secret-change-me", "Auth secret key for cookie signing")
	var config Config
	if err := env.Parse(&config); err == nil {
		return &config
	}
	return &Config{
		SecretKey: cmdSecretKey,
	}
}
