package shortener

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServiceConfig struct {
	BaseURL *string `env:"BASE_URL,required"`
}

func Init() *ServiceConfig {
	cmdBaseURL := flag.String("b", "http://localhost:8080", "server server port")
	var cfg ServiceConfig
	if err := env.Parse(&cfg); err == nil {
		return &cfg
	}
	return &ServiceConfig{
		BaseURL: cmdBaseURL,
	}
}
