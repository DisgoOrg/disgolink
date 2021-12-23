package lavalink

import (
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/log"
	"net/http"
)

type Config struct {
	Logger     log.Logger
	HTTPClient *http.Client
	UserID     discord.Snowflake
	Plugins    []interface{}
}

type ConfigOpt func(config *Config)

func (c *Config) Apply(opts []ConfigOpt) {
	for _, opt := range opts {
		opt(c)
	}
}

// WithLogger lets you inject your own logger implementing log.Logger
//goland:noinspection GoUnusedExportedFunction
func WithLogger(logger log.Logger) ConfigOpt {
	return func(config *Config) {
		config.Logger = logger
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHTTPClient(httpClient *http.Client) ConfigOpt {
	return func(config *Config) {
		config.HTTPClient = httpClient
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithUserID(userID discord.Snowflake) ConfigOpt {
	return func(config *Config) {
		config.UserID = userID
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithPlugins(plugins ...interface{}) ConfigOpt {
	return func(config *Config) {
		config.Plugins = append(config.Plugins, plugins...)
	}
}
