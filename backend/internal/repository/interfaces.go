package repository

import (
	"context"

	"event-manager/internal/models"
)

// EventRepository defines the interface for event data operations
type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, eventID string) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	UpdateStatus(ctx context.Context, eventID string, isActive bool) error
	List(ctx context.Context, limit int) ([]*models.Event, error)
}

// InteractionRepository defines the interface for interaction data operations
type InteractionRepository interface {
	// Create creates a new interaction and returns the generated ID
	Create(ctx context.Context, eventID string, interaction *models.Interaction) (string, error)

	// CreateWithID creates an interaction with a specific ID (for VOTE where userID is the doc ID)
	CreateWithID(ctx context.Context, eventID, recordID string, interaction *models.Interaction) error

	// GetByEventID returns all interactions for an event
	GetByEventID(ctx context.Context, eventID string) ([]*models.Interaction, error)

	// GetByUserAndType returns interactions filtered by user and type
	GetByUserAndType(ctx context.Context, eventID, userID string, iType models.InteractionType) ([]*models.Interaction, error)

	// GetByID returns a specific interaction
	GetByID(ctx context.Context, eventID, recordID string) (*models.Interaction, error)

	// Update updates specific fields of an interaction
	Update(ctx context.Context, eventID, recordID string, updates map[string]interface{}) error

	// Delete removes an interaction
	Delete(ctx context.Context, eventID, recordID string) error
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateFields(ctx context.Context, userID string, updates map[string]interface{}) error
	Exists(ctx context.Context, userID string) (bool, error)
}

// Repositories holds all repository instances
type Repositories struct {
	Events       EventRepository
	Interactions InteractionRepository
	Users        UserRepository
	Close        func() error
}
