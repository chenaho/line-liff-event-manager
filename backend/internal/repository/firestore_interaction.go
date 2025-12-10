package repository

import (
	"context"

	"event-manager/internal/models"

	"cloud.google.com/go/firestore"
)

// FirestoreInteractionRepository implements InteractionRepository using Firestore
type FirestoreInteractionRepository struct {
	client *FirestoreClient
}

// NewFirestoreInteractionRepository creates a new FirestoreInteractionRepository
func NewFirestoreInteractionRepository(client *FirestoreClient) *FirestoreInteractionRepository {
	return &FirestoreInteractionRepository{client: client}
}

func (r *FirestoreInteractionRepository) Create(ctx context.Context, eventID string, interaction *models.Interaction) (string, error) {
	ref, _, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Add(ctx, interaction)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func (r *FirestoreInteractionRepository) CreateWithID(ctx context.Context, eventID, recordID string, interaction *models.Interaction) error {
	_, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID).Set(ctx, interaction)
	return err
}

func (r *FirestoreInteractionRepository) GetByEventID(ctx context.Context, eventID string) ([]*models.Interaction, error) {
	docs, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	interactions := make([]*models.Interaction, 0, len(docs))
	for _, doc := range docs {
		var interaction models.Interaction
		if err := doc.DataTo(&interaction); err != nil {
			continue
		}
		// Store document ID for reference
		interaction.ID = doc.Ref.ID
		interactions = append(interactions, &interaction)
	}
	return interactions, nil
}

func (r *FirestoreInteractionRepository) GetByUserAndType(ctx context.Context, eventID, userID string, iType models.InteractionType) ([]*models.Interaction, error) {
	query := r.client.Client.Collection("events").Doc(eventID).Collection("records").
		Where("userId", "==", userID).
		Where("type", "==", iType)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	interactions := make([]*models.Interaction, 0, len(docs))
	for _, doc := range docs {
		var interaction models.Interaction
		if err := doc.DataTo(&interaction); err != nil {
			continue
		}
		interaction.ID = doc.Ref.ID
		interactions = append(interactions, &interaction)
	}
	return interactions, nil
}

func (r *FirestoreInteractionRepository) GetByID(ctx context.Context, eventID, recordID string) (*models.Interaction, error) {
	doc, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var interaction models.Interaction
	if err := doc.DataTo(&interaction); err != nil {
		return nil, err
	}
	interaction.ID = doc.Ref.ID
	return &interaction, nil
}

func (r *FirestoreInteractionRepository) Update(ctx context.Context, eventID, recordID string, updates map[string]interface{}) error {
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: path, Value: value})
	}

	_, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID).Update(ctx, firestoreUpdates)
	return err
}

func (r *FirestoreInteractionRepository) Delete(ctx context.Context, eventID, recordID string) error {
	_, err := r.client.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID).Delete(ctx)
	return err
}

// Transaction support for complex operations
func (r *FirestoreInteractionRepository) RunTransaction(ctx context.Context, fn func(ctx context.Context, tx *firestore.Transaction) error) error {
	return r.client.Client.RunTransaction(ctx, fn)
}

// GetClient returns the underlying Firestore client for advanced operations
func (r *FirestoreInteractionRepository) GetClient() *firestore.Client {
	return r.client.Client
}
