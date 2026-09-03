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
	Token        string `mapstructure:"token"`
	RefreshToken string `mapstructure:"refresh_token"`
}

type API struct {
	BaseURL string `mapstructure:"base_url"`
}

// DefaultBaseURL is set at build time via ldflags, or overridden by FARO_HELM_API_URL env var.
var DefaultBaseURL = "http://localhost:3001"

// configDirOverride, when set, takes precedence over the user's home
// directory. Intended for use by tests — see SetConfigDirForTesting.
var configDirOverride string

// SetConfigDirForTesting overrides the directory used by Load/Save/GetConfigDir/
// GetConfigFile. For use in tests only — call with "" to restore the default.
func SetConfigDirForTesting(dir string) {
	configDirOverride = dir
}

func resolveConfigDir() string {
	if configDirOverride != "" {
		return configDirOverride
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user home directory: %v", err))
	}
	return filepath.Join(homeDir, ".faro-helm")
}

func GetConfigDir() string  { return resolveConfigDir() }
func GetConfigFile() string { return filepath.Join(resolveConfigDir(), "config.yaml") }

func Load() (*Config, error) {
	configDir := resolveConfigDir()
	configFile := GetConfigFile()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	baseURL := DefaultBaseURL
	if envURL := os.Getenv("FARO_HELM_API_URL"); envURL != "" {
		baseURL = envURL
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return &Config{API: &API{BaseURL: baseURL}}, nil
	}

	// viper is a global singleton — reset it so a stale Set() from an
	// earlier Save()/Load() cycle (e.g. across tests using different
	// SetConfigDirForTesting dirs in the same process) can't leak in ahead
	// of what's actually on disk.
	viper.Reset()
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

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(resolveConfigDir(), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// viper is a global singleton — reset it first so a previously-read or
	// previously-written config (e.g. a different SetConfigDirForTesting
	// dir, or another config.Load() earlier in this process) can't leak
	// into what gets written here.
	viper.Reset()

	// viper.WriteConfigAs serializes Set() values by lowercasing Go field
	// names — it does not consult the mapstructure tags Unmarshal reads by
	// (e.g. RefreshToken would round-trip as "refreshtoken", not
	// "refresh_token"). Write explicit maps keyed to match those tags.
	viper.Set("user", userToMap(cfg.User))
	viper.Set("organization", organizationToMap(cfg.Organization))
	viper.Set("auth", authToMap(cfg.Auth))
	viper.Set("api", nil)

	if err := viper.WriteConfigAs(GetConfigFile()); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func authToMap(a *Auth) map[string]any {
	if a == nil {
		return nil
	}
	return map[string]any{
		"token":         a.Token,
		"refresh_token": a.RefreshToken,
	}
}

func userToMap(u *User) map[string]any {
	if u == nil {
		return nil
	}
	return map[string]any{
		"id":         u.ID,
		"account_id": u.AccountID,
		"email":      u.Email,
		"name":       u.Name,
		"role":       u.Role,
	}
}

func organizationToMap(o *Organization) map[string]any {
	if o == nil {
		return nil
	}
	return map[string]any{
		"id":     o.ID,
		"name":   o.Name,
		"status": o.Status,
	}
}

func (c *Config) IsAuthenticated() bool {
	return c.Auth != nil && c.Auth.Token != ""
}

func (c *Config) Clear() {
	c.User = nil
	c.Organization = nil
	c.Auth = nil
}

func (c *Config) SetAuthData(accessToken, refreshToken string, user *User, org *Organization) {
	c.Auth = &Auth{Token: accessToken, RefreshToken: refreshToken}
	c.User = user
	c.Organization = org
}
