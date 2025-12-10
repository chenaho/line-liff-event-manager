package service

import (
	"context"
	"errors"
	"log"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"cloud.google.com/go/firestore"
)

// InteractionService handles interaction business logic
// Note: This service requires transaction support which is database-specific
// For now, it works with FirestoreInteractionRepository which exposes GetClient()
type InteractionService struct {
	Repo   *repository.FirestoreInteractionRepository
	Events repository.EventRepository
	Users  repository.UserRepository
	Cache  *CacheService
}

func NewInteractionService(repo *repository.FirestoreInteractionRepository, events repository.EventRepository, users repository.UserRepository, cache *CacheService) *InteractionService {
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
	return s.Repo.CreateWithID(ctx, eventID, action.UserID, action)
}

func (s *InteractionService) handleLineUp(ctx context.Context, eventID string, action *models.Interaction) error {
	client := s.Repo.GetClient()

	return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		eventRef := client.Collection("events").Doc(eventID)
		eventDoc, err := tx.Get(eventRef)
		if err != nil {
			return err
		}

		var event models.Event
		if err := eventDoc.DataTo(&event); err != nil {
			return err
		}

		if !event.IsActive {
			return errors.New("event is not active")
		}

		recordsRef := eventRef.Collection("records")

		if action.Count > 0 {
			// +1 Registration
			userRef := client.Collection("users").Doc(action.UserID)
			userDoc, err := tx.Get(userRef)
			isAdmin := false
			if err == nil {
				var user models.User
				userDoc.DataTo(&user)
				isAdmin = user.Role == "admin"
			}

			// Read all records to count
			recordsSnap, err := tx.Documents(recordsRef).GetAll()
			if err != nil {
				return err
			}

			userActiveCount := 0
			totalActiveCount := 0

			for _, doc := range recordsSnap {
				var rec models.Interaction
				doc.DataTo(&rec)
				if rec.Type == models.InteractionTypeLineUp && rec.Status != "CANCELLED" {
					totalActiveCount++
					if rec.UserID == action.UserID {
						userActiveCount++
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
					waitlistCount := 0
					for _, doc := range recordsSnap {
						var rec models.Interaction
						doc.DataTo(&rec)
						if rec.Type == models.InteractionTypeLineUp && rec.Status == "WAITLIST" {
							waitlistCount++
						}
					}

					if waitlistCount >= event.Config.WaitlistLimit {
						return errors.New("waitlist is full")
					}
				}
				action.Status = "WAITLIST"
			} else {
				action.Status = "SUCCESS"
			}

			action.Timestamp = time.Now()
			return tx.Create(recordsRef.NewDoc(), action)

		} else if action.Count < 0 {
			// -1 Cancellation (LIFO)
			recordsSnap, err := tx.Documents(recordsRef).GetAll()
			if err != nil {
				return err
			}

			var latestRecord *firestore.DocumentSnapshot
			var latestTime time.Time

			for _, doc := range recordsSnap {
				var rec models.Interaction
				doc.DataTo(&rec)
				if rec.UserID == action.UserID &&
					rec.Type == models.InteractionTypeLineUp &&
					rec.Status != "CANCELLED" {
					if latestRecord == nil || rec.Timestamp.After(latestTime) {
						latestRecord = doc
						latestTime = rec.Timestamp
					}
				}
			}

			if latestRecord == nil {
				return errors.New("no active registration found")
			}

			now := time.Now()
			return tx.Update(latestRecord.Ref, []firestore.Update{
				{Path: "status", Value: "CANCELLED"},
				{Path: "cancelledAt", Value: now},
			})
		}

		return errors.New("invalid count value")
	})
}

func (s *InteractionService) handleMemo(ctx context.Context, eventID string, action *models.Interaction) error {
	client := s.Repo.GetClient()
	recordsRef := client.Collection("events").Doc(eventID).Collection("records")

	// Check count
	q := recordsRef.Where("userId", "==", action.UserID).Where("type", "==", models.InteractionTypeMemo)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	event, err := s.Events.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if len(docs) >= event.Config.MaxCommentsPerUser {
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
	client := s.Repo.GetClient()
	recordRef := client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		if record.UserID != userID {
			return errors.New("unauthorized: can only edit own registration")
		}

		return tx.Update(recordRef, []firestore.Update{
			{Path: "note", Value: note},
		})
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}

func (s *InteractionService) UpdateMemoContent(ctx context.Context, eventID, recordID, userID, content string) error {
	client := s.Repo.GetClient()
	recordRef := client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		if record.UserID != userID {
			return errors.New("unauthorized: can only edit own message")
		}

		return tx.Update(recordRef, []firestore.Update{
			{Path: "content", Value: content},
		})
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}

func (s *InteractionService) IncrementClapCount(ctx context.Context, eventID, recordID string) error {
	client := s.Repo.GetClient()
	recordRef := client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		newCount := record.ClapCount + 1
		if newCount > 99 {
			newCount = 99
		}

		return tx.Update(recordRef, []firestore.Update{
			{Path: "clapCount", Value: newCount},
		})
	})

	if err == nil {
		s.Cache.Invalidate(eventID)
	}

	return err
}
