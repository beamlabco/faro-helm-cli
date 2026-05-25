package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the Faro Helm CLI configuration
type Config struct {
	User         *User         `mapstructure:"user"`
	Organization *Organization `mapstructure:"organization"`
	Auth         *Auth         `mapstructure:"auth"`
	API          *API          `mapstructure:"api"`
}

type User struct {
	ID        string `mapstructure:"id"`
	AccountID string `mapstructure:"account_id"`
	Email     string `mapstructure:"email"`
	Name      string `mapstructure:"name"`
	Role      string `mapstructure:"role"`
}

type Organization struct {
	ID     string `mapstructure:"id"`
	Name   string `mapstructure:"name"`
	Status string `mapstructure:"status"`
}

type Auth struct {
	Token string `mapstructure:"token"`
}

type API struct {
	BaseURL     string `mapstructure:"base_url"`
	AuthBaseURL string `mapstructure:"auth_base_url"`
}

// DefaultBaseURL is set at build time via ldflags, or overridden by FARO_HELM_API_URL env var.
var DefaultBaseURL = "http://localhost:3001"

// DefaultAuthBaseURL is set at build time via ldflags, or overridden by FARO_AUTH_API_URL env var.
var DefaultAuthBaseURL = "http://localhost:3000"

var (
	configDir  string
	configFile string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user home directory: %v", err))
	}
	configDir = filepath.Join(homeDir, ".faro-helm")
	configFile = filepath.Join(configDir, "config.yaml")
}

func GetConfigDir() string  { return configDir }
func GetConfigFile() string { return configFile }

func Load() (*Config, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	baseURL := DefaultBaseURL
	if envURL := os.Getenv("FARO_HELM_API_URL"); envURL != "" {
		baseURL = envURL
	}

	authBaseURL := DefaultAuthBaseURL
	if envURL := os.Getenv("FARO_AUTH_API_URL"); envURL != "" {
		authBaseURL = envURL
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return &Config{API: &API{BaseURL: baseURL, AuthBaseURL: authBaseURL}}, nil
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	viper.SetDefault("api.base_url", baseURL)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.API == nil {
		cfg.API = &API{}
	}
	cfg.API.BaseURL = baseURL
	cfg.API.AuthBaseURL = authBaseURL

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("user", cfg.User)
	viper.Set("organization", cfg.Organization)
	viper.Set("auth", cfg.Auth)
	viper.Set("api", nil)

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) IsAuthenticated() bool {
	return c.Auth != nil && c.Auth.Token != ""
}

func (c *Config) Clear() {
	c.User = nil
	c.Organization = nil
	c.Auth = nil
}

func (c *Config) SetAuthData(token string, user *User, org *Organization) {
	c.Auth = &Auth{Token: token}
	c.User = user
	c.Organization = org
}
