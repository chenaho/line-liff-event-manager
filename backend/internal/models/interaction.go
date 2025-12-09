package models

import "time"

type InteractionType string

const (
	InteractionTypeVote   InteractionType = "VOTE"
	InteractionTypeLineUp InteractionType = "LINEUP"
	InteractionTypeMemo   InteractionType = "MEMO"
)

type Interaction struct {
	UserID          string          `json:"userId" firestore:"userId"`
	UserDisplayName string          `json:"userDisplayName" firestore:"userDisplayName"`
	UserPictureUrl  string          `json:"userPictureUrl,omitempty" firestore:"userPictureUrl,omitempty"`
	Type            InteractionType `json:"type" firestore:"type"`
	Timestamp       time.Time       `json:"timestamp" firestore:"timestamp"`

	// VOTE
	SelectedOptions []string `json:"selectedOptions,omitempty" firestore:"selectedOptions,omitempty"`

	// LINEUP
	Count       int        `json:"count,omitempty" firestore:"count,omitempty"`   // 1 or -1
	Status      string     `json:"status,omitempty" firestore:"status,omitempty"` // SUCCESS | WAITLIST | CANCELLED
	Note        string     `json:"note,omitempty" firestore:"note,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty" firestore:"cancelledAt,omitempty"` // Timestamp when cancelled (soft delete)

	// MEMO
	Content   string   `json:"content,omitempty" firestore:"content,omitempty"`
	ClapCount int      `json:"clapCount,omitempty" firestore:"clapCount,omitempty"` // Clap reactions count (max 99)
	Reactions []string `json:"reactions,omitempty" firestore:"reactions,omitempty"`
}
