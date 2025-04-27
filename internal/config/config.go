package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	AppName     string     `yaml:"app_name" env-required:"true"`
	Env         string     `yaml:"env" env-required:"true" env-default:"prod"`
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
	User     string `yaml:"user" env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
}

const (
	flagConfigPathName = "config"
	envConfigPathName  = "CONFIG_PATH"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

func MustLoad() *Config {
	cfgPath := fetchConfigPath()
	if cfgPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic("config file not found: " + cfgPath)
	}

	config := &Config{}
	if err := cleanenv.ReadConfig(cfgPath, config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return config
}

func fetchConfigPath() string {
	var path string
	flag.StringVar(&path, flagConfigPathName, "", "path ot config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv(envConfigPathName)
	}

	return path
}
