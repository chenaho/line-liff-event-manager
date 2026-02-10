package repository

import (
	"context"
	"database/sql"
	"time"
)

// PostgresSettingsRepository implements SettingsRepository for PostgreSQL
type PostgresSettingsRepository struct {
	client *PostgresClient
}

// NewPostgresSettingsRepository creates a new PostgresSettingsRepository
func NewPostgresSettingsRepository(client *PostgresClient) *PostgresSettingsRepository {
	return &PostgresSettingsRepository{client: client}
}

// Get retrieves a setting value by key
func (r *PostgresSettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.client.DB.QueryRowContext(ctx,
		"SELECT value FROM settings WHERE key = $1", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// Set upserts a setting value
func (r *PostgresSettingsRepository) Set(ctx context.Context, key, value string) error {
	_, err := r.client.DB.ExecContext(ctx,
		`INSERT INTO settings (key, value, updated_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3`,
		key, value, time.Now())
	return err
}

// GetAll retrieves all settings
func (r *PostgresSettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.client.DB.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, rows.Err()
}
