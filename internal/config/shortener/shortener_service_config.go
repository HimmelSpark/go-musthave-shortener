package shortener

import "flag"

type ServiceConfig struct {
	BaseURL *string
}

func Init() *ServiceConfig {
	return &ServiceConfig{
		BaseURL: flag.String("b", "http://localhost:8080", "server server port"),
	}
}
