package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	AppName string `envconfig:"NAME" required:"true"`
	Env     string `envconfig:"ENV" required:"true"`
	Debug   bool   `envconfig:"DEBUG" default:"false"`
	DB      DB
}

type DB struct {
	Host     string `envconfig:"POSTGRES_HOST" required:"true"`
	Port     string `envconfig:"POSTGRES_PORT" required:"true"`
	Name     string `envconfig:"POSTGRES_DB" required:"true"`
	Username string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
}

func MustLoad() *Config {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	return &cfg
}
