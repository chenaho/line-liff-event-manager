package service

import (
	"context"
	"log"

	"event-manager/internal/repository"
)

// SettingsService handles settings business logic
type SettingsService struct {
	Repo repository.SettingsRepository
}

// NewSettingsService creates a new SettingsService
func NewSettingsService(repo repository.SettingsRepository) *SettingsService {
	return &SettingsService{Repo: repo}
}

// Get retrieves a setting value by key
func (s *SettingsService) Get(ctx context.Context, key string) (string, error) {
	return s.Repo.Get(ctx, key)
}

// Set upserts a setting value
func (s *SettingsService) Set(ctx context.Context, key, value string) error {
	log.Printf("[Settings] Setting %s", key)
	return s.Repo.Set(ctx, key, value)
}

// GetAll retrieves all settings
func (s *SettingsService) GetAll(ctx context.Context) (map[string]string, error) {
	return s.Repo.GetAll(ctx)
}
