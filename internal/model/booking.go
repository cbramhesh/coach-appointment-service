package model

import "time"

type Booking struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	CoachID   int       `json:"coach_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
