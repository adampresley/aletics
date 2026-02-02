package configuration

import (
	"github.com/adampresley/configinator"
	"github.com/adampresley/mux"
)

type Config struct {
	mux.Config

	CookieSecret string `flag:"cookiesecret" env:"COOKIE_SECRET" default:"password" description:"Secret for encoding coodies"`
	DSN          string `flag:"dsn" env:"DSN" default:"file:./aletics.db" description:"Database connection"`
	LogLevel     string `flag:"loglevel" env:"LOG_LEVEL" default:"debug" description:"The log level to use. Valid values are 'debug', 'info', 'warn', and 'error'"`
	PageSize     int    `flag:"pagesize" env:"PAGE_SIZE" default:"10" description:"The number of items to display per page"`
}

func LoadConfig() Config {
	config := Config{}
	configinator.Behold(&config)
	return config
}
