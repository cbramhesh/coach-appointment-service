package dto

type CreateBookingRequest struct {
	UserID    int    `json:"user_id"    validate:"required,gt=0"`
	CoachID   int    `json:"coach_id"   validate:"required,gt=0"`
	StartTime string `json:"start_time" validate:"required"`
}

type BookingResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CoachID   int    `json:"coach_id"`
	CoachName string `json:"coach_name,omitempty"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CancelBookingRequest struct {
	UserID int `json:"user_id" validate:"required,gt=0"`
}
