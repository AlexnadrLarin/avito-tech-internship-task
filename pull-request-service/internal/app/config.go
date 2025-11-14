package app

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig   
}

type ServerConfig struct {
	Port            int    
	Host            string 
	ShutdownTimeout int    
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	loadEnvVars(config)
	return config, nil
}

func loadEnvVars(config *Config) {
	if envVal := os.Getenv("PR_REQUEST_HOST"); envVal != "" {
		config.Server.Host = envVal
	}
	if envVal := os.Getenv("PR_REQUEST_PORT"); envVal != "" {
		if port, err := strconv.Atoi(envVal); err == nil {
			config.Server.Port = port
		}
	}
	if envVal := os.Getenv("PR_REQUEST_SERVER_SHUTDOWN_TIMEOUT"); envVal != "" {
		if timeout, err := strconv.Atoi(envVal); err == nil {
			config.Server.ShutdownTimeout = timeout
		}
	}
}
