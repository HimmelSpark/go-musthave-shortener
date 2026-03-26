package storage

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func Init() (*Config, error) {
	cmdPath := flag.String("f", "urls.json", "File storage path")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if cfg.FileStoragePath != "" {
		return &cfg, nil
	}
	return &Config{
		FileStoragePath: *cmdPath,
	}, nil
}
