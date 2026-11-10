package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Level string

const (
	Dev  Level = "dev"
	Prod Level = "prod"
)

type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	App      AppConfig
}

type PostgresConfig struct {
	DB   string `env:"POSTGRES_DB,required"`
	User string `env:"POSTGRES_USER,required"`
	Pass string `env:"POSTGRES_PASSWORD,required"`
	Port string `env:"POSTGRES_PORT,required"`
	Host string `env:"POSTGRES_HOST,required"`
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST,required"`
	Port string `env:"REDIS_PORT,required"`
	User string `env:"REDIS_USERNAME,required"`
	Pass string `env:"REDIS_PASSWORD,required"`
	DB   int    `env:"REDIS_DB,required"`
}

type AppConfig struct {
	Port       string `env:"PORT"`
	DebugLevel Level  `env:"DEBUG_LEVEL"`
}

func NewLoad() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, fmt.Errorf("postgres env.Parse: %w", err)
	}

	if err := env.Parse(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("redis env.Parse: %w", err)
	}

	if err := env.Parse(&cfg.App); err != nil {
		return nil, fmt.Errorf("app env.Parse: %w", err)
	}

	level := Level(cfg.App.DebugLevel)
	if level != Dev && level != Prod {
		return nil, fmt.Errorf("incorrect debug level: %s", level)
	}
	return &cfg, nil
}
