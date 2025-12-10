package repository

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// FirestoreClient wraps the Firestore client
type FirestoreClient struct {
	Client *firestore.Client
}

// NewFirestoreClient creates a new Firestore client
func NewFirestoreClient() (*FirestoreClient, error) {
	ctx := context.Background()
	credsPath := os.Getenv("FIREBASE_CREDENTIALS")

	var app *firebase.App
	var err error

	if credsPath != "" {
		opt := option.WithCredentialsFile(credsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return &FirestoreClient{Client: client}, nil
}

// Close closes the Firestore client
func (c *FirestoreClient) Close() error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
