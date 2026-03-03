package server

import "flag"

type Config struct {
	ServerAddress *string
}

func Init() *Config {
	return &Config{
		ServerAddress: flag.String("a", "localhost:8080", "HTTP server address"),
	}
}
