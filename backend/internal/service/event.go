package service

import (
	"context"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type EventService struct {
	Repo *repository.FirestoreRepository
}

func NewEventService(repo *repository.FirestoreRepository) *EventService {
	return &EventService{Repo: repo}
}

func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	event.EventID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.IsActive = true

	_, err := s.Repo.Client.Collection("events").Doc(event.EventID).Set(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, eventID string) (*models.Event, error) {
	doc, err := s.Repo.Client.Collection("events").Doc(eventID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := doc.DataTo(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *EventService) UpdateEventStatus(ctx context.Context, eventID string, isActive bool) error {
	_, err := s.Repo.Client.Collection("events").Doc(eventID).Update(ctx, []firestore.Update{
		{Path: "isActive", Value: isActive},
	})
	return err
}

func (s *EventService) UpdateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	// Get existing event to preserve createdAt and createdBy
	existingEvent, err := s.GetEvent(ctx, event.EventID)
	if err != nil {
		return nil, err
	}

	// Preserve original creation metadata
	event.CreatedAt = existingEvent.CreatedAt
	event.CreatedBy = existingEvent.CreatedBy

	// Update the event
	_, err = s.Repo.Client.Collection("events").Doc(event.EventID).Set(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) ListEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	iter := s.Repo.Client.Collection("events").OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
	// Initialize as empty slice instead of nil to return [] instead of null in JSON
	events := make([]*models.Event, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var event models.Event
		if err := doc.DataTo(&event); err != nil {
			continue
		}
		events = append(events, &event)
	}
	return events, nil
}
