package repository

import (
	"errors"
	"log"
	"os"
)

// Config holds database configuration
type Config struct {
	Postgres PostgresConfig // PostgreSQL configuration
}

// PostgresConfig holds PostgreSQL connection settings
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// LoadConfigFromEnv creates a Config from environment variables
func LoadConfigFromEnv() *Config {
	cfg := &Config{}

	// Load PostgreSQL config
	cfg.Postgres = PostgresConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "eventmanager"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: getEnvOrDefault("POSTGRES_DB", "eventmanager"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
	}

	log.Printf("[Config] PostgreSQL: host=%s, port=%s, user=%s, db=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Database)

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewRepositories creates PostgreSQL repository instances
func NewRepositories(cfg *Config) (*Repositories, error) {
	if cfg.Postgres.Password == "" {
		return nil, errors.New("POSTGRES_PASSWORD is required")
	}

	client, err := NewPostgresClient(&cfg.Postgres)
	if err != nil {
		return nil, err
	}

	log.Printf("[Repositories] PostgreSQL connection established")

	return &Repositories{
		Events:       NewPostgresEventRepository(client),
		Interactions: NewPostgresInteractionRepository(client),
		Users:        NewPostgresUserRepository(client),
		Close: func() error {
			return client.Close()
		},
	}, nil
}
