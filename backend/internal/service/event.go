package service

import (
	"context"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"github.com/google/uuid"
)

type EventService struct {
	Repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) *EventService {
	return &EventService{Repo: repo}
}

func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	event.EventID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.IsActive = true

	if err := s.Repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, eventID string) (*models.Event, error) {
	return s.Repo.GetByID(ctx, eventID)
}

func (s *EventService) UpdateEventStatus(ctx context.Context, eventID string, isActive bool) error {
	return s.Repo.UpdateStatus(ctx, eventID, isActive)
}

func (s *EventService) UpdateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	// Get existing event to preserve createdAt and createdBy
	existingEvent, err := s.Repo.GetByID(ctx, event.EventID)
	if err != nil {
		return nil, err
	}

	// Preserve original creation metadata
	event.CreatedAt = existingEvent.CreatedAt
	event.CreatedBy = existingEvent.CreatedBy

	// Update the event
	if err := s.Repo.Update(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) ListEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	return s.Repo.List(ctx, limit)
}
