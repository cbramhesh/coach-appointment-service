package repository

import (
	"context"
	"database/sql"
	"time"

	"coach-appointment-service/internal/model"
)

type BookingRepo struct {
	db *sql.DB
}

func NewBookingRepo(db *sql.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) GetConfirmedByCoachAndRange(ctx context.Context, coachID int, from, to time.Time) ([]model.Booking, error) {
	return nil, nil
}
