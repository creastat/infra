package config

import (
	"time"
)

// BaseConfig provides common configuration fields for all services
type BaseConfig struct {
	Server        ServerConfig        `yaml:"server" json:"server"`
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int           `yaml:"port" json:"port"`
	Host         string        `yaml:"host" json:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	Logging LoggingConfig `yaml:"logging" json:"logging"`
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`
	Tracing TracingConfig `yaml:"tracing" json:"tracing"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`   // trace, debug, info, warn, error
	Format string `yaml:"format" json:"format"` // json, text
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled  bool    `yaml:"enabled" json:"enabled"`
	Endpoint string  `yaml:"endpoint" json:"endpoint"`
	Sampler  float64 `yaml:"sampler" json:"sampler"` // Sampling rate 0.0-1.0
}

// SetServerDefaults sets default values for server configuration
func (c *BaseConfig) SetServerDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 60 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 60 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
}

// SetObservabilityDefaults sets default values for observability configuration
func (c *BaseConfig) SetObservabilityDefaults() {
	if c.Observability.Logging.Level == "" {
		c.Observability.Logging.Level = "info"
	}
	if c.Observability.Logging.Format == "" {
		c.Observability.Logging.Format = "json"
	}
	if c.Observability.Metrics.Port == 0 {
		c.Observability.Metrics.Port = 9090
	}
	if c.Observability.Metrics.Path == "" {
		c.Observability.Metrics.Path = "/metrics"
	}
	if c.Observability.Tracing.Sampler == 0 {
		c.Observability.Tracing.Sampler = 1.0
	}
}
