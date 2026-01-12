package models

import "time"

type EventType string

const (
	EventTypeVote   EventType = "VOTE"
	EventTypeLineUp EventType = "LINEUP"
	EventTypeMemo   EventType = "MEMO"
)

type EventConfig struct {
	// VOTE
	MaxVotes   int      `json:"maxVotes,omitempty" firestore:"maxVotes,omitempty"`     // 1 = single select, >1 = multi-select
	ShowVoters *bool    `json:"showVoters,omitempty" firestore:"showVoters,omitempty"` // true = show voters, false = anonymous (pointer to distinguish unset from false)
	Options    []string `json:"options,omitempty" firestore:"options,omitempty"`

	// LINEUP
	MaxParticipants int       `json:"maxParticipants,omitempty" firestore:"maxParticipants,omitempty"`
	WaitlistLimit   int       `json:"waitlistLimit,omitempty" firestore:"waitlistLimit,omitempty"`
	MaxCountPerUser int       `json:"maxCountPerUser,omitempty" firestore:"maxCountPerUser,omitempty"`
	PrivacyMode     bool      `json:"privacyMode,omitempty" firestore:"privacyMode,omitempty"` // Hide full name and use avatar
	StartTime       time.Time `json:"startTime,omitempty" firestore:"startTime,omitempty"`
	EndTime         time.Time `json:"endTime,omitempty" firestore:"endTime,omitempty"`

	// MEMO
	MaxCommentsPerUser int  `json:"maxCommentsPerUser,omitempty" firestore:"maxCommentsPerUser,omitempty"`
	AllowReaction      bool `json:"allowReaction,omitempty" firestore:"allowReaction,omitempty"`
}

type Event struct {
	EventID    string      `json:"eventId" firestore:"eventId"`
	Type       EventType   `json:"type" firestore:"type"`
	Title      string      `json:"title" firestore:"title"`
	IsActive   bool        `json:"isActive" firestore:"isActive"`
	IsArchived bool        `json:"isArchived" firestore:"isArchived"`
	CreatedBy  string      `json:"createdBy" firestore:"createdBy"`
	CreatedAt  time.Time   `json:"createdAt" firestore:"createdAt"`
	Config     EventConfig `json:"config" firestore:"config"`
}
