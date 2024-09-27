package config

import "fmt"

// Config holds the server configuration
type Config struct {
	Host string
	Port int
}

// Load loads the configuration from environment variables or a config file
func Load() (*Config, error) {
	// TODO: Implement actual configuration loading
	// For now, return a default configuration
	return &Config{
		Host: "localhost",
		Port: 6667,
	}, nil
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
