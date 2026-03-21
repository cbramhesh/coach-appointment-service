package repository

import (
	"coach-appointment-service/internal/model"
	"context"
	"database/sql"
)

type CoachRepo struct {
	db *sql.DB
}

func NewCoachRepo(db *sql.DB) *CoachRepo {
	return &CoachRepo{db: db}
}

func (r *CoachRepo) GetByID(ctx context.Context, id int) (*model.Coach, error) {
	coach := &model.Coach{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, timezone, created_at, updated_at FROM coaches WHERE id = ?",
		id,
	).Scan(&coach.ID, &coach.Name, &coach.Timezone, &coach.CreatedAt, &coach.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return coach, nil
}
