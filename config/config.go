package config

import (
	"flag"
	"fmt"
)

// VerbosityLevel represents the logging verbosity level
type VerbosityLevel int

const (
	Info VerbosityLevel = iota
	Debug
	Trace
)

// Config holds the server configuration
type Config struct {
	Host      string
	Port      int
	Verbosity VerbosityLevel
}

// Load loads the configuration from command-line flags
func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Host, "host", "localhost", "Host to listen on")
	flag.IntVar(&cfg.Port, "port", 6667, "Port to listen on")
	verbosity := flag.String("verbosity", "info", "Logging verbosity (info, debug, trace)")

	flag.Parse()

	switch *verbosity {
	case "info":
		cfg.Verbosity = Info
	case "debug":
		cfg.Verbosity = Debug
	case "trace":
		cfg.Verbosity = Trace
	default:
		return nil, fmt.Errorf("invalid verbosity level: %s", *verbosity)
	}

	return cfg, nil
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
