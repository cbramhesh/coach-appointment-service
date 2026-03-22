package model

import "time"

type Availability struct {
	ID        int       `json:"id"`
	CoachID   int       `json:"coach_id"`
	DayOfWeek int       `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
}
