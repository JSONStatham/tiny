package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName     string `envconfig:"NAME" required:"true"`
	Env         string `envconfig:"ENV" default:"prod"`
	Debug       bool   `envconfig:"DEBUG" default:"false"`
	AliasLength int    `envconfig:"ALIAS_LENGTH" default:"6"`
	HttpServer  HttpServer
	DB          DB
	MsgBroker   MsgBroker
	Cache       Cache
}

type HttpServer struct {
	Address     string        `envconfig:"HTTP_ADDRESS" required:"true"`
	TimeOut     time.Duration `envconfig:"HTTP_TIMEOUT"`
	IdleTimeout time.Duration `envconfig:"HTTP_IDLE_TIMEOUT"`
	User        string        `envconfig:"BASIC_AUTH_USER"`
	Password    string        `envconfig:"BASIC_AUTH_PASSWORD"`
}

type DB struct {
	Host     string `envconfig:"POSTGRES_HOST" required:"true"`
	Port     string `envconfig:"POSTGRES_PORT" required:"true"`
	Name     string `envconfig:"POSTGRES_DB" required:"true"`
	Username string `envconfig:"POSTGRES_USER" required:"true"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
}

type MsgBroker struct {
	Addr         []string `envconfig:"KAFKA_ADDRESS" required:"true"`
	FlushTimeout int      `envconfig:"KAFKA_PRODUCER_FLUSH_TIME" default:"5000"`
}

type Cache struct {
	Addr     string `envconfig:"REDIS_ADDRESS"`
	Password string `envconfig:"REDIS_PASSWORD"`
	Db       int    `envconfig:"REDIS_DB"`
}

func MustLoad() *Config {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	return &cfg
}
