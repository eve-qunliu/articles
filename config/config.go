package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DBHost     string `envconfig:"POSTGRES_HOST"`
	DBName     string `envconfig:"POSTGRES_DB"`
	DBUser     string `envconfig:"POSTGRES_USER"`
	DBPassword string `envconfig:"POSTGRES_PASSWORD"`
	TagLimit   int    `envconfig:"TAG_LIMIT"`
}

func NewConfig() *Config {
	cfg := &Config{}
	envconfig.MustProcess("", cfg)
	return cfg
}
