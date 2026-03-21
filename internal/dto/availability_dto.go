package dto

type CreateAvailabilityRequest struct {
	CoachID   int    `json:"coach_id"    validate:"required,gt=0"`
	DayOfWeek int    `json:"day_of_week" validate:"gte=0,lte=6"`
	StartTime string `json:"start_time"  validate:"required"`
	EndTime   string `json:"end_time"    validate:"required"`
}
