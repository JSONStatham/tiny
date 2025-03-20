package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppName     string     `yaml:"app_name" env-required:"true"`
	Env         string     `yaml:"env" env-required:"true"`
	Debug       bool       `yaml:"debug" env-default:"false"`
	AliasLength int        `yaml:"alias_length"`
	HttpServer  HttpServer `yaml:"http_server"`
	DB          DB         `yaml:"db"`
}

type HttpServer struct {
	Address     string        `yaml:"address"`
	TimeOut     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	User        string        `env:"BASIC_AUTH_USER"`
	Password    string        `env:"BASIC_AUTH_PASSWORD"`
}

type DB struct {
	Host     string `yaml:"host" env:"DB_HOST"`
	Port     string `yaml:"port" env:"DB_PORT"`
	Name     string `yaml:"name" env:"DB_DATABASE"`
	User     string `yaml:"user" env:"DB_USERNAME"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
}

func MustLoad() Config {
	var cfg Config

	cfgPath := fetchConfigPath()
	if _, err := os.Stat(cfgPath); err != nil {
		panic("error opening config file")
	}

	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		panic("Failed to read config")
	}

	return cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config_path", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
