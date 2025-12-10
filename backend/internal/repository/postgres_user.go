package repository

import (
	"context"
	"database/sql"

	"event-manager/internal/models"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	client *PostgresClient
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(client *PostgresClient) *PostgresUserRepository {
	return &PostgresUserRepository{client: client}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (line_user_id, line_display_name, picture_url, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.client.DB.ExecContext(ctx, query,
		user.LineUserID, user.LineDisplayName, user.PictureURL, user.Role, user.CreatedAt)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT line_user_id, line_display_name, picture_url, role, created_at
		FROM users WHERE line_user_id = $1
	`
	var user models.User
	var displayName, pictureUrl sql.NullString

	err := r.client.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.LineUserID, &displayName, &pictureUrl, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	if displayName.Valid {
		user.LineDisplayName = displayName.String
	}
	if pictureUrl.Valid {
		user.PictureURL = pictureUrl.String
	}

	return &user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET line_display_name = $2, picture_url = $3, role = $4
		WHERE line_user_id = $1
	`
	_, err := r.client.DB.ExecContext(ctx, query,
		user.LineUserID, user.LineDisplayName, user.PictureURL, user.Role)
	return err
}

func (r *PostgresUserRepository) UpdateFields(ctx context.Context, userID string, updates map[string]interface{}) error {
	// Simple implementation: fetch, update, save
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	for key, value := range updates {
		switch key {
		case "lineDisplayName":
			user.LineDisplayName = value.(string)
		case "pictureUrl":
			user.PictureURL = value.(string)
		case "role":
			user.Role = value.(string)
		}
	}

	return r.Update(ctx, user)
}

func (r *PostgresUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE line_user_id = $1)`
	var exists bool
	err := r.client.DB.QueryRowContext(ctx, query, userID).Scan(&exists)
	return exists, err
}
