package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"event-manager/internal/models"

	_ "github.com/lib/pq"
)

// PostgresClient holds the PostgreSQL connection
type PostgresClient struct {
	DB *sql.DB
}

// NewPostgresClient creates a new PostgreSQL client
func NewPostgresClient(cfg *PostgresConfig) (*PostgresClient, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &PostgresClient{DB: db}, nil
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// PostgresEventRepository implements EventRepository using PostgreSQL
type PostgresEventRepository struct {
	client *PostgresClient
}

// NewPostgresEventRepository creates a new PostgresEventRepository
func NewPostgresEventRepository(client *PostgresClient) *PostgresEventRepository {
	return &PostgresEventRepository{client: client}
}

func (r *PostgresEventRepository) Create(ctx context.Context, event *models.Event) error {
	configJSON, err := json.Marshal(event.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO events (event_id, type, title, is_active, created_by, created_at, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.client.DB.ExecContext(ctx, query,
		event.EventID, event.Type, event.Title, event.IsActive, event.CreatedBy, event.CreatedAt, configJSON)
	return err
}

func (r *PostgresEventRepository) GetByID(ctx context.Context, eventID string) (*models.Event, error) {
	query := `
		SELECT event_id, type, title, is_active, created_by, created_at, config
		FROM events WHERE event_id = $1
	`
	var event models.Event
	var configJSON []byte

	err := r.client.DB.QueryRowContext(ctx, query, eventID).Scan(
		&event.EventID, &event.Type, &event.Title, &event.IsActive, &event.CreatedBy, &event.CreatedAt, &configJSON)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &event.Config); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *PostgresEventRepository) Update(ctx context.Context, event *models.Event) error {
	configJSON, err := json.Marshal(event.Config)
	if err != nil {
		return err
	}

	query := `
		UPDATE events 
		SET type = $2, title = $3, is_active = $4, config = $5
		WHERE event_id = $1
	`
	_, err = r.client.DB.ExecContext(ctx, query,
		event.EventID, event.Type, event.Title, event.IsActive, configJSON)
	return err
}

func (r *PostgresEventRepository) UpdateStatus(ctx context.Context, eventID string, isActive bool) error {
	query := `UPDATE events SET is_active = $2 WHERE event_id = $1`
	_, err := r.client.DB.ExecContext(ctx, query, eventID, isActive)
	return err
}

func (r *PostgresEventRepository) List(ctx context.Context, limit int) ([]*models.Event, error) {
	query := `
		SELECT event_id, type, title, is_active, created_by, created_at, config
		FROM events ORDER BY created_at DESC LIMIT $1
	`
	rows, err := r.client.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*models.Event, 0)
	for rows.Next() {
		var event models.Event
		var configJSON []byte

		if err := rows.Scan(&event.EventID, &event.Type, &event.Title, &event.IsActive, &event.CreatedBy, &event.CreatedAt, &configJSON); err != nil {
			continue
		}
		if err := json.Unmarshal(configJSON, &event.Config); err != nil {
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}
