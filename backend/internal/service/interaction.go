package service

import (
	"context"
	"errors"
	"log"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"
)

// InteractionService handles interaction business logic
type InteractionService struct {
	Repo   repository.InteractionRepository
	Events repository.EventRepository
	Users  repository.UserRepository
	Cache  *CacheService
}

// NewInteractionService creates an InteractionService with repository
func NewInteractionService(repo repository.InteractionRepository, events repository.EventRepository, users repository.UserRepository, cache *CacheService) *InteractionService {
	return &InteractionService{
		Repo:   repo,
		Events: events,
		Users:  users,
		Cache:  cache,
	}
}

func (s *InteractionService) HandleAction(ctx context.Context, eventID string, action *models.Interaction) error {
	action.Timestamp = time.Now()

	var err error
	switch action.Type {
	case models.InteractionTypeVote:
		err = s.handleVote(ctx, eventID, action)
	case models.InteractionTypeLineUp:
		err = s.handleLineUp(ctx, eventID, action)
	case models.InteractionTypeMemo:
		err = s.handleMemo(ctx, eventID, action)
	default:
		return errors.New("unknown action type")
	}

	// Invalidate cache after successful write
	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}

func (s *InteractionService) handleVote(ctx context.Context, eventID string, action *models.Interaction) error {
	// Use composite ID: eventID_userID to ensure one vote per user per event
	recordID := eventID + "_" + action.UserID
	return s.Repo.CreateWithID(ctx, eventID, recordID, action)
}

func (s *InteractionService) handleLineUp(ctx context.Context, eventID string, action *models.Interaction) error {
	// Get event
	event, err := s.Events.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !event.IsActive {
		return errors.New("event is not active")
	}

	// Get all interactions for this event
	allInteractions, err := s.Repo.GetByEventID(ctx, eventID)
	if err != nil {
		return err
	}

	// Check if user is admin
	user, _ := s.Users.GetByID(ctx, action.UserID)
	isAdmin := user != nil && user.Role == "admin"

	if action.Count > 0 {
		// +1 Registration
		userActiveCount := 0
		totalActiveCount := 0
		waitlistCount := 0

		for _, rec := range allInteractions {
			if rec.Type == models.InteractionTypeLineUp && rec.Status != "CANCELLED" {
				totalActiveCount++
				if rec.UserID == action.UserID {
					userActiveCount++
				}
				if rec.Status == "WAITLIST" {
					waitlistCount++
				}
			}
		}

		// Check user limit (admin bypasses)
		maxPerUser := event.Config.MaxCountPerUser
		if !isAdmin && maxPerUser > 0 && userActiveCount >= maxPerUser {
			return errors.New("registration limit reached")
		}

		// Determine status based on capacity
		if totalActiveCount >= event.Config.MaxParticipants {
			if event.Config.WaitlistLimit > 0 {
				if waitlistCount >= event.Config.WaitlistLimit {
					return errors.New("waitlist is full")
				}
			}
			action.Status = "WAITLIST"
		} else {
			action.Status = "SUCCESS"
		}

		action.Timestamp = time.Now()
		_, err := s.Repo.Create(ctx, eventID, action)
		return err

	} else if action.Count < 0 {
		// -1 Cancellation (LIFO - Last In, First Out)
		var latestRecord *models.Interaction
		var latestTime time.Time

		for _, rec := range allInteractions {
			if rec.UserID == action.UserID &&
				rec.Type == models.InteractionTypeLineUp &&
				rec.Status != "CANCELLED" {
				if latestRecord == nil || rec.Timestamp.After(latestTime) {
					latestRecord = rec
					latestTime = rec.Timestamp
				}
			}
		}

		if latestRecord == nil {
			return errors.New("no active registration found")
		}

		now := time.Now()
		return s.Repo.Update(ctx, eventID, latestRecord.ID, map[string]interface{}{
			"status":      "CANCELLED",
			"cancelledAt": now,
		})
	}

	return errors.New("invalid count value")
}

func (s *InteractionService) handleMemo(ctx context.Context, eventID string, action *models.Interaction) error {
	// Get user's memo count
	userMemos, err := s.Repo.GetByUserAndType(ctx, eventID, action.UserID, models.InteractionTypeMemo)
	if err != nil {
		return err
	}

	// Get event config
	event, err := s.Events.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if len(userMemos) >= event.Config.MaxCommentsPerUser {
		return errors.New("max comments reached")
	}

	_, err = s.Repo.Create(ctx, eventID, action)
	return err
}

func (s *InteractionService) GetEventStatus(ctx context.Context, eventID string) (map[string]interface{}, error) {
	log.Printf("[GetEventStatus] Fetching status for event: %s", eventID)

	// Check cache first
	if cached, found := s.Cache.Get(eventID); found {
		log.Printf("[GetEventStatus] Cache HIT for event: %s", eventID)
		return cached, nil
	}
	log.Printf("[GetEventStatus] Cache MISS for event: %s", eventID)

	// Fetch all records
	interactions, err := s.Repo.GetByEventID(ctx, eventID)
	if err != nil {
		log.Printf("[GetEventStatus] ERROR getting records: %v", err)
		return nil, err
	}

	log.Printf("[GetEventStatus] Successfully fetched %d total records", len(interactions))

	// Build result
	result := make(map[string]interface{})
	list := make([]map[string]interface{}, 0, len(interactions))

	for _, rec := range interactions {
		recMap := map[string]interface{}{
			"id":              rec.ID,
			"type":            rec.Type,
			"userId":          rec.UserID,
			"userDisplayName": rec.UserDisplayName,
			"userPictureUrl":  rec.UserPictureUrl,
			"timestamp":       rec.Timestamp,
			"status":          rec.Status,
			"selectedOptions": rec.SelectedOptions,
			"count":           rec.Count,
			"note":            rec.Note,
			"content":         rec.Content,
			"clapCount":       rec.ClapCount,
		}
		list = append(list, recMap)
	}

	log.Printf("[GetEventStatus] Returning %d records for event: %s", len(list), eventID)

	result["records"] = list

	// Cache the result
	s.Cache.Set(eventID, result)

	return result, nil
}

func (s *InteractionService) UpdateRegistrationNote(ctx context.Context, eventID, recordID, userID, note string) error {
	// Get the record to verify ownership
	record, err := s.Repo.GetByID(ctx, eventID, recordID)
	if err != nil {
		return errors.New("record not found")
	}

	if record.UserID != userID {
		return errors.New("unauthorized: can only edit own registration")
	}

	err = s.Repo.Update(ctx, eventID, recordID, map[string]interface{}{
		"note": note,
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}

func (s *InteractionService) UpdateMemoContent(ctx context.Context, eventID, recordID, userID, content string) error {
	// Get the record to verify ownership
	record, err := s.Repo.GetByID(ctx, eventID, recordID)
	if err != nil {
		return errors.New("record not found")
	}

	if record.UserID != userID {
		return errors.New("unauthorized: can only edit own message")
	}

	err = s.Repo.Update(ctx, eventID, recordID, map[string]interface{}{
		"content": content,
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}

func (s *InteractionService) IncrementClapCount(ctx context.Context, eventID, recordID string) error {
	// Get current clap count
	record, err := s.Repo.GetByID(ctx, eventID, recordID)
	if err != nil {
		return errors.New("record not found")
	}

	newCount := record.ClapCount + 1
	if newCount > 200 {
		newCount = 200
	}

	err = s.Repo.Update(ctx, eventID, recordID, map[string]interface{}{
		"clapCount": newCount,
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}
