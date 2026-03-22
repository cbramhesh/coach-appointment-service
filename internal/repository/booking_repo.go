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

func (r *BookingRepo) Create(ctx context.Context, b *model.Booking) (*model.Booking, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO bookings (user_id, coach_id, start_time, end_time, status)
		 VALUES (?, ?, ?, ?, 'confirmed')`,
		b.UserID, b.CoachID, b.StartTime, b.EndTime,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	b.ID = int(id)
	b.Status = "confirmed"
	return b, nil
}

func (r *BookingRepo) GetConfirmedByCoachAndRange(ctx context.Context, coachID int, from, to time.Time) ([]model.Booking, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, coach_id, start_time, end_time, status, created_at
		 FROM bookings
		 WHERE coach_id = ? AND start_time >= ? AND start_time < ? AND status = 'confirmed'
		 ORDER BY start_time`,
		coachID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.CoachID,
			&b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *BookingRepo) GetUpcomingByUser(ctx context.Context, userID int) ([]model.Booking, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, coach_id, start_time, end_time, status, created_at
		 FROM bookings
		 WHERE user_id = ? AND start_time > NOW() AND status = 'confirmed'
		 ORDER BY start_time`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.CoachID,
			&b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *BookingRepo) GetByID(ctx context.Context, id int) (*model.Booking, error) {
	b := &model.Booking{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, coach_id, start_time, end_time, status, created_at
		 FROM bookings WHERE id = ?`, id,
	).Scan(&b.ID, &b.UserID, &b.CoachID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt)
	return b, err
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE bookings SET status = ? WHERE id = ?", status, id)
	return err
}
