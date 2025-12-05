package repository

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type FirestoreRepository struct {
	Client *firestore.Client
}

func NewFirestoreRepository() (*FirestoreRepository, error) {
	ctx := context.Background()
	credsPath := os.Getenv("FIREBASE_CREDENTIALS")

	var app *firebase.App
	var err error

	if credsPath != "" {
		opt := option.WithCredentialsFile(credsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		// Fallback for when running in GCP environment (automatic auth)
		// or if just testing without explicit key file (might fail later)
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return &FirestoreRepository{Client: client}, nil
}

func (r *FirestoreRepository) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}
