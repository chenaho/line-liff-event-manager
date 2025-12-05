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
	Type            InteractionType `json:"type" firestore:"type"`
	Timestamp       time.Time       `json:"timestamp" firestore:"timestamp"`

	// VOTE
	SelectedOptions []string `json:"selectedOptions,omitempty" firestore:"selectedOptions,omitempty"`

	// LINEUP
	Count  int    `json:"count,omitempty" firestore:"count,omitempty"`   // 1 or -1
	Status string `json:"status,omitempty" firestore:"status,omitempty"` // SUCCESS | WAITLIST
	Note   string `json:"note,omitempty" firestore:"note,omitempty"`

	// MEMO
	Content   string   `json:"content,omitempty" firestore:"content,omitempty"`
	Reactions []string `json:"reactions,omitempty" firestore:"reactions,omitempty"`
}
