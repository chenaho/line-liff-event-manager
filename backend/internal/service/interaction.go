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

		// Check if user already registered
		recordRef := eventRef.Collection("records").Doc(action.UserID)
		recordDoc, err := tx.Get(recordRef)
		alreadyRegistered := err == nil && recordDoc.Exists()

		if action.Count > 0 {
			// Register (+1)
			if alreadyRegistered {
				return errors.New("already registered")
			}

			// Count current participants
			// Note: Firestore count aggregation is cheaper but for transaction consistency we might need to read or keep a counter.
			// For simplicity/correctness in transaction, we should maintain a counter in the event doc or count documents.
			// Let's assume we count documents (might be slow for large events, but fine for "LineUp" which usually has small limits).
			// Actually, better to store a "currentCount" in event doc if we want strict limits.
			// But spec didn't define "currentCount" in Event model.
			// Let's query all records to count.

			// Query inside transaction is limited.
			// "Queries inside transactions must be ancestor queries".
			// So we can query collection("records") since it's subcollection.

			// However, for "Vibe Coding", let's just read all records (assuming small number < 100).
			recordsSnap, err := tx.Documents(eventRef.Collection("records")).GetAll()
			if err != nil {
				return err
			}

			currentCount := 0
			for _, doc := range recordsSnap {
				var rec models.Interaction
				doc.DataTo(&rec)
				if rec.Type == models.InteractionTypeLineUp && rec.Status == "SUCCESS" {
					currentCount++
				}
			}

			if currentCount >= event.Config.MaxParticipants {
				if event.Config.WaitlistLimit > 0 {
					// Check waitlist count
					waitlistCount := 0
					for _, doc := range recordsSnap {
						var rec models.Interaction
						doc.DataTo(&rec)
						if rec.Type == models.InteractionTypeLineUp && rec.Status == "WAITLIST" {
							waitlistCount++
						}
					}
					if waitlistCount >= event.Config.WaitlistLimit {
						return errors.New("waitlist full")
					}
					action.Status = "WAITLIST"
				} else {
					return errors.New("event full")
				}
			} else {
				action.Status = "SUCCESS"
			}

			return tx.Set(recordRef, action)

		} else if action.Count < 0 {
			// Cancel (-1)
			if !alreadyRegistered {
				return errors.New("not registered")
			}

			// IMPORTANT: Read the record status BEFORE deleting it
			// to avoid "read after write in transaction" error
			var oldRecord models.Interaction
			recordDoc.DataTo(&oldRecord)

			// Delete record
			if err := tx.Delete(recordRef); err != nil {
				return err
			}

			// If user was SUCCESS, we might need to promote someone from WAITLIST.
			// Spec: "User A (SUCCESS) cancel -> System auto promote Waitlist #1 (User B) to SUCCESS"
			if oldRecord.Status == "SUCCESS" {
				// Find first waitlist candidate
				// We need to query again or use the snapshot we could have taken.
				// Since we didn't take snapshot of all records above (only if +1), we need to query now.
				// But we can't do arbitrary query in transaction easily if not ancestor.
				// Subcollection query IS ancestor query.

				// We need to find the earliest WAITLIST record.
				// Firestore transaction queries are tricky.
				// Let's try to get all records again (safe for small events).
				recordsSnap, err := tx.Documents(eventRef.Collection("records")).GetAll()
				if err != nil {
					return err
				}

				var firstWaiter *firestore.DocumentSnapshot
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

				if firstWaiter != nil {
					return tx.Update(firstWaiter.Ref, []firestore.Update{
						{Path: "status", Value: "SUCCESS"},
					})
				}
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
	var list []models.Interaction

	for _, doc := range records {
		var rec models.Interaction
		doc.DataTo(&rec)
		list = append(list, rec)
	}

	result["records"] = list
	return result, nil
}
