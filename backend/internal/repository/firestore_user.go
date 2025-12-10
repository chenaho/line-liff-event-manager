package repository

import (
	"context"
	"strings"

	"event-manager/internal/models"

	"cloud.google.com/go/firestore"
)

// FirestoreUserRepository implements UserRepository using Firestore
type FirestoreUserRepository struct {
	client *FirestoreClient
}

// NewFirestoreUserRepository creates a new FirestoreUserRepository
func NewFirestoreUserRepository(client *FirestoreClient) *FirestoreUserRepository {
	return &FirestoreUserRepository{client: client}
}

func (r *FirestoreUserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.client.Client.Collection("users").Doc(user.LineUserID).Set(ctx, user)
	return err
}

func (r *FirestoreUserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	doc, err := r.client.Client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *FirestoreUserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.client.Client.Collection("users").Doc(user.LineUserID).Set(ctx, user)
	return err
}

func (r *FirestoreUserRepository) UpdateFields(ctx context.Context, userID string, updates map[string]interface{}) error {
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: path, Value: value})
	}

	_, err := r.client.Client.Collection("users").Doc(userID).Update(ctx, firestoreUpdates)
	return err
}

func (r *FirestoreUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	doc, err := r.client.Client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return doc.Exists(), nil
}
