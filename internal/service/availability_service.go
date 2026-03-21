package service

import (
	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/dto"
	"coach-appointment-service/internal/model"
	"coach-appointment-service/internal/repository"
	"context"
	"fmt"
	"strings"
	"time"
)

type AvailabilityService struct {
	coachRepo *repository.CoachRepo
	availRepo *repository.AvailabilityRepo
}

func NewAvailabilityService(ar *repository.AvailabilityRepo, cr *repository.CoachRepo) *AvailabilityService {
	return &AvailabilityService{
		coachRepo: cr,
		availRepo: ar,
	}
}

func (s *AvailabilityService) CreateAvailability(ctx context.Context, req dto.CreateAvailabilityRequest) (*model.Availability, error) {
	_, err := s.coachRepo.GetByID(ctx, req.CoachID)
	if err != nil {
		return nil, apperror.NotFound("COACH_NOT_FOUND",
			fmt.Sprintf("Coach with ID %d not found", req.CoachID))
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, apperror.BadRequest("INVALID_TIME_FORMAT", "start_time must be HH:mm format")
	}
	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, apperror.BadRequest("INVALID_TIME_FORMAT", "end_time must be HH:mm format")
	}

	if !endTime.After(startTime) {
		return nil, apperror.BadRequest("INVALID_TIME_RANGE", "end_time must be after start_time")
	}
	if endTime.Sub(startTime) < 30*time.Minute {
		return nil, apperror.BadRequest("INVALID_TIME_RANGE", "Window must be at least 30 minutes")
	}

	avail := &model.Availability{
		CoachID:   req.CoachID,
		DayOfWeek: req.DayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
	}

	result, err := s.availRepo.Create(ctx, avail)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate enrty") {
			return nil, apperror.Conflict("DUPLICATE_AVAILABILITY", "this availability window already exists")
		}
		return nil, err
	}

	return result, nil
}
