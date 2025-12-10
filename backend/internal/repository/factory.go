package repository

import (
	"errors"
	"os"
)

// Config holds database configuration
type Config struct {
	Type     string         // "firestore" or "postgres"
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
	cfg := &Config{
		Type: os.Getenv("DB_TYPE"),
	}

	// Default to firestore if not specified
	if cfg.Type == "" {
		cfg.Type = "firestore"
	}

	// Load PostgreSQL config
	cfg.Postgres = PostgresConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "eventmanager"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: getEnvOrDefault("POSTGRES_DB", "eventmanager"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewRepositories creates repository instances based on configuration
func NewRepositories(cfg *Config) (*Repositories, error) {
	switch cfg.Type {
	case "firestore":
		return newFirestoreRepositories()
	case "postgres":
		return newPostgresRepositories(&cfg.Postgres)
	default:
		return nil, errors.New("unknown database type: " + cfg.Type)
	}
}

// newFirestoreRepositories creates Firestore-backed repositories
func newFirestoreRepositories() (*Repositories, error) {
	client, err := NewFirestoreClient()
	if err != nil {
		return nil, err
	}

	return &Repositories{
		Events:       NewFirestoreEventRepository(client),
		Interactions: NewFirestoreInteractionRepository(client),
		Users:        NewFirestoreUserRepository(client),
		Close: func() error {
			return client.Close()
		},
	}, nil
}

// newPostgresRepositories creates PostgreSQL-backed repositories
func newPostgresRepositories(cfg *PostgresConfig) (*Repositories, error) {
	client, err := NewPostgresClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		Events:       NewPostgresEventRepository(client),
		Interactions: NewPostgresInteractionRepository(client),
		Users:        NewPostgresUserRepository(client),
		Close: func() error {
			return client.Close()
		},
	}, nil
}
