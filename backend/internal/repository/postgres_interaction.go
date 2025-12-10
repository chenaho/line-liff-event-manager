package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"event-manager/internal/models"
)

// PostgresInteractionRepository implements InteractionRepository using PostgreSQL
type PostgresInteractionRepository struct {
	client *PostgresClient
}

// NewPostgresInteractionRepository creates a new PostgresInteractionRepository
func NewPostgresInteractionRepository(client *PostgresClient) *PostgresInteractionRepository {
	return &PostgresInteractionRepository{client: client}
}

// interactionPayload holds the JSONB payload fields
type interactionPayload struct {
	SelectedOptions []string `json:"selectedOptions,omitempty"`
	Count           int      `json:"count,omitempty"`
	Note            string   `json:"note,omitempty"`
	CancelledAt     *string  `json:"cancelledAt,omitempty"`
	Content         string   `json:"content,omitempty"`
	ClapCount       int      `json:"clapCount,omitempty"`
	Reactions       []string `json:"reactions,omitempty"`
}

func (r *PostgresInteractionRepository) Create(ctx context.Context, eventID string, interaction *models.Interaction) (string, error) {
	payload := interactionPayload{
		SelectedOptions: interaction.SelectedOptions,
		Count:           interaction.Count,
		Note:            interaction.Note,
		Content:         interaction.Content,
		ClapCount:       interaction.ClapCount,
		Reactions:       interaction.Reactions,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	var id string
	query := `
		INSERT INTO interactions (event_id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = r.client.DB.QueryRowContext(ctx, query,
		eventID, interaction.UserID, interaction.Type, interaction.UserDisplayName,
		interaction.UserPictureUrl, interaction.Status, interaction.Timestamp, payloadJSON).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *PostgresInteractionRepository) CreateWithID(ctx context.Context, eventID, recordID string, interaction *models.Interaction) error {
	payload := interactionPayload{
		SelectedOptions: interaction.SelectedOptions,
		Count:           interaction.Count,
		Note:            interaction.Note,
		Content:         interaction.Content,
		ClapCount:       interaction.ClapCount,
		Reactions:       interaction.Reactions,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO interactions (id, event_id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_display_name = EXCLUDED.user_display_name,
			user_picture_url = EXCLUDED.user_picture_url,
			status = EXCLUDED.status,
			timestamp = EXCLUDED.timestamp,
			payload = EXCLUDED.payload
	`
	_, err = r.client.DB.ExecContext(ctx, query,
		recordID, eventID, interaction.UserID, interaction.Type, interaction.UserDisplayName,
		interaction.UserPictureUrl, interaction.Status, interaction.Timestamp, payloadJSON)
	return err
}

func (r *PostgresInteractionRepository) GetByEventID(ctx context.Context, eventID string) ([]*models.Interaction, error) {
	query := `
		SELECT id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload
		FROM interactions WHERE event_id = $1 ORDER BY timestamp ASC
	`
	rows, err := r.client.DB.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInteractions(rows)
}

func (r *PostgresInteractionRepository) GetByUserAndType(ctx context.Context, eventID, userID string, iType models.InteractionType) ([]*models.Interaction, error) {
	query := `
		SELECT id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload
		FROM interactions WHERE event_id = $1 AND user_id = $2 AND type = $3 ORDER BY timestamp ASC
	`
	rows, err := r.client.DB.QueryContext(ctx, query, eventID, userID, iType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanInteractions(rows)
}

func (r *PostgresInteractionRepository) GetByID(ctx context.Context, eventID, recordID string) (*models.Interaction, error) {
	query := `
		SELECT id, user_id, type, user_display_name, user_picture_url, status, timestamp, payload
		FROM interactions WHERE event_id = $1 AND id = $2
	`
	var interaction models.Interaction
	var payloadJSON []byte
	var displayName, pictureUrl, status sql.NullString

	err := r.client.DB.QueryRowContext(ctx, query, eventID, recordID).Scan(
		&interaction.ID, &interaction.UserID, &interaction.Type, &displayName, &pictureUrl, &status, &interaction.Timestamp, &payloadJSON)
	if err != nil {
		return nil, err
	}

	if displayName.Valid {
		interaction.UserDisplayName = displayName.String
	}
	if pictureUrl.Valid {
		interaction.UserPictureUrl = pictureUrl.String
	}
	if status.Valid {
		interaction.Status = status.String
	}

	var payload interactionPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, err
	}

	interaction.SelectedOptions = payload.SelectedOptions
	interaction.Count = payload.Count
	interaction.Note = payload.Note
	interaction.Content = payload.Content
	interaction.ClapCount = payload.ClapCount
	interaction.Reactions = payload.Reactions

	return &interaction, nil
}

func (r *PostgresInteractionRepository) Update(ctx context.Context, eventID, recordID string, updates map[string]interface{}) error {
	// For simplicity, fetch current, merge, and update
	current, err := r.GetByID(ctx, eventID, recordID)
	if err != nil {
		return err
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "status":
			current.Status = value.(string)
		case "note":
			current.Note = value.(string)
		case "content":
			current.Content = value.(string)
		case "clapCount":
			current.ClapCount = value.(int)
		}
	}

	payload := interactionPayload{
		SelectedOptions: current.SelectedOptions,
		Count:           current.Count,
		Note:            current.Note,
		Content:         current.Content,
		ClapCount:       current.ClapCount,
		Reactions:       current.Reactions,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `UPDATE interactions SET status = $3, payload = $4 WHERE event_id = $1 AND id = $2`
	_, err = r.client.DB.ExecContext(ctx, query, eventID, recordID, current.Status, payloadJSON)
	return err
}

func (r *PostgresInteractionRepository) Delete(ctx context.Context, eventID, recordID string) error {
	query := `DELETE FROM interactions WHERE event_id = $1 AND id = $2`
	_, err := r.client.DB.ExecContext(ctx, query, eventID, recordID)
	return err
}

func (r *PostgresInteractionRepository) scanInteractions(rows *sql.Rows) ([]*models.Interaction, error) {
	interactions := make([]*models.Interaction, 0)

	for rows.Next() {
		var interaction models.Interaction
		var payloadJSON []byte
		var displayName, pictureUrl, status sql.NullString

		if err := rows.Scan(&interaction.ID, &interaction.UserID, &interaction.Type, &displayName, &pictureUrl, &status, &interaction.Timestamp, &payloadJSON); err != nil {
			continue
		}

		if displayName.Valid {
			interaction.UserDisplayName = displayName.String
		}
		if pictureUrl.Valid {
			interaction.UserPictureUrl = pictureUrl.String
		}
		if status.Valid {
			interaction.Status = status.String
		}

		var payload interactionPayload
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			continue
		}

		interaction.SelectedOptions = payload.SelectedOptions
		interaction.Count = payload.Count
		interaction.Note = payload.Note
		interaction.Content = payload.Content
		interaction.ClapCount = payload.ClapCount
		interaction.Reactions = payload.Reactions

		interactions = append(interactions, &interaction)
	}

	return interactions, nil
}
