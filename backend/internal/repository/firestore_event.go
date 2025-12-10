package repository

import (
	"context"

	"event-manager/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FirestoreEventRepository implements EventRepository using Firestore
type FirestoreEventRepository struct {
	client *FirestoreClient
}

// NewFirestoreEventRepository creates a new FirestoreEventRepository
func NewFirestoreEventRepository(client *FirestoreClient) *FirestoreEventRepository {
	return &FirestoreEventRepository{client: client}
}

func (r *FirestoreEventRepository) Create(ctx context.Context, event *models.Event) error {
	_, err := r.client.Client.Collection("events").Doc(event.EventID).Set(ctx, event)
	return err
}

func (r *FirestoreEventRepository) GetByID(ctx context.Context, eventID string) (*models.Event, error) {
	doc, err := r.client.Client.Collection("events").Doc(eventID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := doc.DataTo(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *FirestoreEventRepository) Update(ctx context.Context, event *models.Event) error {
	_, err := r.client.Client.Collection("events").Doc(event.EventID).Set(ctx, event)
	return err
}

func (r *FirestoreEventRepository) UpdateStatus(ctx context.Context, eventID string, isActive bool) error {
	_, err := r.client.Client.Collection("events").Doc(eventID).Update(ctx, []firestore.Update{
		{Path: "isActive", Value: isActive},
	})
	return err
}

func (r *FirestoreEventRepository) UpdateArchived(ctx context.Context, eventID string, isArchived bool) error {
	_, err := r.client.Client.Collection("events").Doc(eventID).Update(ctx, []firestore.Update{
		{Path: "isArchived", Value: isArchived},
	})
	return err
}

func (r *FirestoreEventRepository) List(ctx context.Context, limit int) ([]*models.Event, error) {
	iter := r.client.Client.Collection("events").OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
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
