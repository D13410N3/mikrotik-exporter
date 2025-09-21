package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Auths   map[string]AuthConfig   `yaml:"auths"`
	Modules map[string]ModuleConfig `yaml:"modules"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// ModuleConfig represents module configuration
type ModuleConfig struct {
	Collectors map[string]bool `yaml:"collectors"`
}

// LoadConfig loads configuration from the specified file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetAuth returns the authentication configuration for the given name
func (c *Config) GetAuth(name string) (AuthConfig, error) {
	auth, exists := c.Auths[name]
	if !exists {
		return AuthConfig{}, fmt.Errorf("auth configuration '%s' not found", name)
	}
	return auth, nil
}

// GetModule returns the module configuration for the given name
func (c *Config) GetModule(name string) (ModuleConfig, error) {
	module, exists := c.Modules[name]
	if !exists {
		return ModuleConfig{}, fmt.Errorf("module configuration '%s' not found", name)
	}
	return module, nil
}
