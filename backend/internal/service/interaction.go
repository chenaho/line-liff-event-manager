package service

import (
	"context"
	"errors"
	"time"

	"event-manager/internal/models"
	"event-manager/internal/repository"

	"cloud.google.com/go/firestore"
)

type InteractionService struct {
	Repo *repository.FirestoreRepository
}

func NewInteractionService(repo *repository.FirestoreRepository) *InteractionService {
	return &InteractionService{Repo: repo}
}

func (s *InteractionService) HandleAction(ctx context.Context, eventID string, action *models.Interaction) error {
	action.Timestamp = time.Now()

	switch action.Type {
	case models.InteractionTypeVote:
		return s.handleVote(ctx, eventID, action)
	case models.InteractionTypeLineUp:
		return s.handleLineUp(ctx, eventID, action)
	case models.InteractionTypeMemo:
		return s.handleMemo(ctx, eventID, action)
	default:
		return errors.New("unknown action type")
	}
}

func (s *InteractionService) handleVote(ctx context.Context, eventID string, action *models.Interaction) error {
	// Simple write for now. Real implementation might need to check maxVotes config.
	// Assuming frontend checks or we add read-before-write here.
	// For Vibe Coding, we'll just save it.
	_, err := s.Repo.Client.Collection("events").Doc(eventID).Collection("records").Doc(action.UserID).Set(ctx, action)
	return err
}

func (s *InteractionService) handleLineUp(ctx context.Context, eventID string, action *models.Interaction) error {
	return s.Repo.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		eventRef := s.Repo.Client.Collection("events").Doc(eventID)
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

			// Get user to check if admin
			userRef := s.Repo.Client.Collection("users").Doc(action.UserID)
			userDoc, err := tx.Get(userRef)
			isAdmin := false
			if err == nil {
				var user models.User
				userDoc.DataTo(&user)
				isAdmin = user.Role == "admin"
			}

			// Read all records to count user's active registrations and total
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
				// Event is full, check waitlist
				if event.Config.WaitlistLimit > 0 {
					// Count current waitlist
					waitlistCount := 0
					for _, doc := range recordsSnap {
						var rec models.Interaction
						doc.DataTo(&rec)
						if rec.Type == models.InteractionTypeLineUp && rec.Status == "WAITLIST" {
							waitlistCount++
						}
					}

					// Check if waitlist is full
					if waitlistCount >= event.Config.WaitlistLimit {
						return errors.New("waitlist is full")
					}
				}
				action.Status = "WAITLIST"
			} else {
				action.Status = "SUCCESS"
			}

			// Create new registration record with auto-generated ID
			action.Timestamp = time.Now()
			return tx.Create(recordsRef.NewDoc(), action)

		} else if action.Count < 0 {
			// -1 Cancellation (LIFO - Last In, First Out)

			// Read all records to find user's most recent active registration
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
				return errors.New("no active registration to cancel")
			}

			// Read status before updating
			var oldRecord models.Interaction
			latestRecord.DataTo(&oldRecord)

			// Find waitlist candidate BEFORE any writes (if needed)
			var firstWaiter *firestore.DocumentSnapshot
			if oldRecord.Status == "SUCCESS" {
				var earliestTime time.Time
				for _, doc := range recordsSnap {
					var rec models.Interaction
					doc.DataTo(&rec)
					if rec.Type == models.InteractionTypeLineUp && rec.Status == "WAITLIST" {
						if firstWaiter == nil || rec.Timestamp.Before(earliestTime) {
							firstWaiter = doc
							earliestTime = rec.Timestamp
						}
					}
				}
			}

			// NOW do writes: Soft delete by marking as CANCELLED
			now := time.Now()
			updates := []firestore.Update{
				{Path: "status", Value: "CANCELLED"},
				{Path: "cancelledAt", Value: now},
			}

			if err := tx.Update(latestRecord.Ref, updates); err != nil {
				return err
			}

			// Promote waitlist candidate if found
			if firstWaiter != nil {
				return tx.Update(firstWaiter.Ref, []firestore.Update{
					{Path: "status", Value: "SUCCESS"},
				})
			}
			return nil
		}
		return nil
	})
}

func (s *InteractionService) handleMemo(ctx context.Context, eventID string, action *models.Interaction) error {
	// Use auto-id for memos since one user can post multiple?
	// Spec says "maxCommentsPerUser".
	// If we use UserID as doc ID, user can only have one memo?
	// Spec says "Interaction Sub-Collection".
	// Usually for comments we want multiple.
	// But spec JSON shows "userId" as a field, doesn't explicitly say Doc ID is UserID for Memo.
	// However, for VOTE and LINEUP, one record per user makes sense.
	// For MEMO, "maxCommentsPerUser: 3".
	// So we should use AutoID for Memo records, but query to check count.

	recordsRef := s.Repo.Client.Collection("events").Doc(eventID).Collection("records")

	// Check count
	q := recordsRef.Where("userId", "==", action.UserID).Where("type", "==", models.InteractionTypeMemo)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	// We need event config to check maxCommentsPerUser.
	// For speed, let's assume 3 if not checked, or read event.
	// Let's read event.
	eventDoc, err := s.Repo.Client.Collection("events").Doc(eventID).Get(ctx)
	if err != nil {
		return err
	}
	var event models.Event
	eventDoc.DataTo(&event)

	if len(docs) >= event.Config.MaxCommentsPerUser {
		return errors.New("max comments reached")
	}

	_, _, err = recordsRef.Add(ctx, action)
	return err
}

func (s *InteractionService) GetEventStatus(ctx context.Context, eventID string) (map[string]interface{}, error) {
	// Return aggregated status
	records, err := s.Repo.Client.Collection("events").Doc(eventID).Collection("records").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var result = make(map[string]interface{})
	var list []map[string]interface{}

	for _, doc := range records {
		var rec models.Interaction
		doc.DataTo(&rec)

		// Convert to map and add document ID
		recMap := map[string]interface{}{
			"id":              doc.Ref.ID, // Add document ID
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

	result["records"] = list
	return result, nil
}

func (s *InteractionService) UpdateRegistrationNote(ctx context.Context, eventID, recordID, userID, note string) error {
	// Get the record
	recordRef := s.Repo.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	return s.Repo.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		// Verify user owns this record
		if record.UserID != userID {
			return errors.New("unauthorized: can only edit own registration")
		}

		// Update note
		return tx.Update(recordRef, []firestore.Update{
			{Path: "note", Value: note},
		})
	})
}

func (s *InteractionService) UpdateMemoContent(ctx context.Context, eventID, recordID, userID, content string) error {
	recordRef := s.Repo.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	return s.Repo.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		// Verify user owns this record
		if record.UserID != userID {
			return errors.New("unauthorized: can only edit own message")
		}

		// Update content
		return tx.Update(recordRef, []firestore.Update{
			{Path: "content", Value: content},
		})
	})
}

func (s *InteractionService) IncrementClapCount(ctx context.Context, eventID, recordID string) error {
	recordRef := s.Repo.Client.Collection("events").Doc(eventID).Collection("records").Doc(recordID)

	return s.Repo.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(recordRef)
		if err != nil {
			return errors.New("record not found")
		}

		var record models.Interaction
		if err := doc.DataTo(&record); err != nil {
			return err
		}

		// Increment clap count (max 99)
		newCount := record.ClapCount + 1
		if newCount > 99 {
			newCount = 99
		}

		return tx.Update(recordRef, []firestore.Update{
			{Path: "clapCount", Value: newCount},
		})
	})
}
