package storage

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	FileStoragePath *string `env:"FILE_STORAGE_PATH,required"`
}

func Init() *Config {
	cmdPath := flag.String("f", "urls.json", "File storage path")
	var cfg Config
	if err := env.Parse(&cfg); err == nil {
		return &cfg
	}
	return &Config{
		FileStoragePath: cmdPath,
	}
}
