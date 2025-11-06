package config

// Config holds application configuration
type Config struct {
	// TODO: Add configuration fields
	Server ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string
	Port int
}

// Load loads configuration from file or environment
func Load(path string) (*Config, error) {
	// TODO: Implement configuration loading
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 50051,
		},
	}, nil
}
