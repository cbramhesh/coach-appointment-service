package repository

import (
	"coach-appointment-service/internal/model"
	"context"
	"database/sql"
)

type AvailabilityRepo struct {
	db *sql.DB
}

func NewAvailabilityRepo(db *sql.DB) *AvailabilityRepo {
	return &AvailabilityRepo{db: db}
}

func (r *AvailabilityRepo) Create(ctx context.Context, a *model.Availability) (*model.Availability, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO availabilities (coach_id, day_of_week, start_time, end_time)
			VALUES (?, ?, ?, ?)`,
		a.CoachID,
		a.DayOfWeek,
		a.StartTime,
		a.EndTime,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	a.ID = int(id)
	return a, nil
}

func (r *AvailabilityRepo) GetByCoachAndDay(ctx context.Context, coachID, dayOfWeek int) ([]model.Availability, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, coach_id, day_of_week, start_time, end_time, created_at
		 FROM availabilities
		 WHERE coach_id = ? AND day_of_week = ?
		 ORDER BY start_time`,
		coachID, dayOfWeek,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []model.Availability
	for rows.Next() {
		var a model.Availability
		if err := rows.Scan(&a.ID, &a.CoachID, &a.DayOfWeek,
			&a.StartTime, &a.EndTime, &a.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, rows.Err()
}
