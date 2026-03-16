package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress *string `env:"SERVER_ADDRESS,required"`
}

func Init() *Config {
	cmdServerAddress := flag.String("a", ":8080", "Server address")
	var config Config
	if err := env.Parse(&config); err == nil {
		return &config
	}
	return &Config{
		ServerAddress: cmdServerAddress,
	}
}
