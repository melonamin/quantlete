package config

import "time"

// Config holds all application configuration.
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Strava  StravaConfig  `mapstructure:"strava"`
	Storage StorageConfig `mapstructure:"storage"`
	Log     LogConfig     `mapstructure:"log"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	DevMode      bool          `mapstructure:"dev_mode"`
}

// StravaConfig holds Strava API configuration.
type StravaConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri"`
}

// StorageConfig holds database configuration.
type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
	DBFile  string `mapstructure:"db_file"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8081,
			Host:         "",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			DevMode:      false,
		},
		Strava: StravaConfig{
			RedirectURI: "http://localhost:8081/api/v1/auth/strava/callback",
		},
		Storage: StorageConfig{
			DataDir: "./data",
			DBFile:  "stata.db",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// DBPath returns the full path to the database file.
func (c *Config) DBPath() string {
	return c.Storage.DataDir + "/" + c.Storage.DBFile
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Strava credentials are optional for initial setup
	// They'll be required when attempting OAuth
	return nil
}
