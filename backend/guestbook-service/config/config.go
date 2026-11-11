package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Postgres PostgresConfig
}

type PostgresConfig struct {
	DB   string `env:"POSTGRES_DB,required"`
	User string `env:"POSTGRES_USER,required"`
	Pass string `env:"POSTGRES_PASSWORD,required"`
	Port string `env:"POSTGRES_PORT,required"`
	Host string `env:"POSTGRES_HOST,required"`
}

func NewLoad() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, fmt.Errorf("postgres env.Parse: %w", err)
	}
	return &cfg, nil
}
