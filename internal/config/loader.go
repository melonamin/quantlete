package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from environment variables and config file.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	defaults := Default()
	v.SetDefault("server.port", defaults.Server.Port)
	v.SetDefault("server.host", defaults.Server.Host)
	v.SetDefault("server.read_timeout", defaults.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", defaults.Server.WriteTimeout)
	v.SetDefault("server.idle_timeout", defaults.Server.IdleTimeout)
	v.SetDefault("server.dev_mode", defaults.Server.DevMode)
	v.SetDefault("strava.redirect_uri", defaults.Strava.RedirectURI)
	v.SetDefault("storage.data_dir", defaults.Storage.DataDir)
	v.SetDefault("storage.db_file", defaults.Storage.DBFile)
	v.SetDefault("log.level", defaults.Log.Level)
	v.SetDefault("log.format", defaults.Log.Format)

	// Environment variables
	v.SetEnvPrefix("STATA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file (optional)
	v.SetConfigName("stata")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.config/stata")
	v.AddConfigPath("/etc/stata")

	// Read config file if it exists (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// LoadWithOverrides loads config and applies command-line overrides.
func LoadWithOverrides(port int, devMode bool) (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Apply command-line overrides
	if port != 0 {
		cfg.Server.Port = port
	}
	cfg.Server.DevMode = devMode

	return cfg, nil
}
