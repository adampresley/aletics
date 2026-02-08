package configuration

import (
	"github.com/adampresley/configinator"
	"github.com/adampresley/mux"
)

type Config struct {
	mux.Config

	CookieSecret     string `flag:"cookiesecret" env:"COOKIE_SECRET" default:"password" description:"Secret for encoding coodies"`
	DSN              string `flag:"dsn" env:"DSN" default:"file:./aletics.db" description:"Database connection"`
	LogLevel         string `flag:"loglevel" env:"LOG_LEVEL" default:"debug" description:"The log level to use. Valid values are 'debug', 'info', 'warn', and 'error'"`
	MaxmindAccountID string `flag:"maxmind-account-id" env:"MAXMIND_ACCOUNT_ID" default:"" description:"MaxMind API account ID"`
	MaxmindApiKey    string `flag:"maxmind-api-key" env:"MAXMIND_API_KEY" default:"" description:"MaxMind API key"`
	PageSize         int    `flag:"pagesize" env:"PAGE_SIZE" default:"10" description:"The number of items to display per page"`
	TLD              string `flag:"tld" env:"TLD" default:"localhost:3000" description:"Top-level domain for this server"`
}

func LoadConfig() Config {
	config := Config{}
	configinator.Behold(&config)
	return config
}
