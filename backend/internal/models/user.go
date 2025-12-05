package models

import "time"

type User struct {
	LineUserID      string    `json:"lineUserId" firestore:"lineUserId"`
	LineDisplayName string    `json:"lineDisplayName" firestore:"lineDisplayName"`
	PictureURL      string    `json:"pictureUrl" firestore:"pictureUrl"`
	CustomName      string    `json:"customName" firestore:"customName"`
	Role            string    `json:"role" firestore:"role"` // "admin" or "user"
	CreatedAt       time.Time `json:"createdAt" firestore:"createdAt"`
}
