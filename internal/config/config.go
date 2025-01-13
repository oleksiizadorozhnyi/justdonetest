package config

import "github.com/caarlos0/env/v9"

var (
	config = Config{}
)

type Config struct {
	ServerPort string `env:"SERVER_PORT" envDefault:":8080"`
	Postgres   Postgres
}

type Postgres struct {
	DBUser     string `env:"DB_USER" envDefault:"user"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"password"`
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBName     string `env:"DB_NAME" envDefault:"orders"`
	SSLMode    string `env:"SSL_MODE" envDefault:"disable"`
}

func LoadConfig() (Config, error) {
	err := env.Parse(&config)
	return config, err
}
