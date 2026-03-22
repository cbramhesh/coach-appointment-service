package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/dto"
	"coach-appointment-service/internal/model"
	"coach-appointment-service/internal/repository"
)

type BookingService struct {
	bookingRepo *repository.BookingRepo
	coachRepo   *repository.CoachRepo
	slotService *SlotService
}

func NewBookingService(br *repository.BookingRepo, cr *repository.CoachRepo, ss *SlotService) *BookingService {
	return &BookingService{bookingRepo: br, coachRepo: cr, slotService: ss}
}

func (s *BookingService) CreateBooking(ctx context.Context, req dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, apperror.BadRequest("INVALID_TIME_FORMAT",
			"start_time must be ISO 8601 format, e.g. 2025-10-28T13:00:00Z")
	}

	if startTime.Before(time.Now().UTC()) {
		return nil, apperror.BadRequest("SLOT_IN_PAST", "Cannot book a slot in the past")
	}

	if startTime.Minute()%30 != 0 || startTime.Second() != 0 {
		return nil, apperror.BadRequest("INVALID_SLOT_BOUNDARY",
			"Slot must start on a :00 or :30 minute mark")
	}

	_, err = s.coachRepo.GetByID(ctx, req.CoachID)
	if err != nil {
		return nil, apperror.NotFound("COACH_NOT_FOUND",
			fmt.Sprintf("Coach with ID %d not found", req.CoachID))
	}

	dateStr := startTime.Format("2006-01-02")
	slotsResp, err := s.slotService.GetAvailableSlots(ctx, req.CoachID, dateStr)
	if err != nil {
		return nil, err
	}

	slotAvailable := false
	for _, slot := range slotsResp.Slots {
		if slot.StartTime == startTime.UTC().Format(time.RFC3339) {
			slotAvailable = true
			break
		}
	}
	if !slotAvailable {
		return nil, apperror.BadRequest("INVALID_SLOT",
			"This time slot is not available for this coach")
	}

	endTime := startTime.Add(30 * time.Minute)
	booking := &model.Booking{
		UserID:    req.UserID,
		CoachID:   req.CoachID,
		StartTime: startTime.UTC(),
		EndTime:   endTime.UTC(),
	}

	result, err := s.bookingRepo.Create(ctx, booking)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, apperror.Conflict("SLOT_ALREADY_BOOKED",
				"This time slot is no longer available")
		}
		return nil, err
	}

	return &dto.BookingResponse{
		ID:        result.ID,
		UserID:    result.UserID,
		CoachID:   result.CoachID,
		StartTime: result.StartTime.Format(time.RFC3339),
		EndTime:   result.EndTime.Format(time.RFC3339),
		Status:    result.Status,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *BookingService) GetUpcomingBookings(ctx context.Context, userID int) ([]dto.BookingResponse, error) {
	bookings, err := s.bookingRepo.GetUpcomingByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		responses = append(responses, dto.BookingResponse{
			ID:        b.ID,
			UserID:    b.UserID,
			CoachID:   b.CoachID,
			StartTime: b.StartTime.Format(time.RFC3339),
			EndTime:   b.EndTime.Format(time.RFC3339),
			Status:    b.Status,
			CreatedAt: b.CreatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID int) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return apperror.NotFound("BOOKING_NOT_FOUND", "Booking not found")
	}
	if booking.UserID != userID {
		return &apperror.AppError{StatusCode: 403, Code: "FORBIDDEN",
			Message: "You can only cancel your own bookings"}
	}
	if booking.Status != "confirmed" {
		return apperror.BadRequest("ALREADY_CANCELLED", "This booking is already cancelled")
	}
	if booking.StartTime.Before(time.Now().UTC()) {
		return apperror.BadRequest("SLOT_IN_PAST", "Cannot cancel a past booking")
	}
	return s.bookingRepo.UpdateStatus(ctx, bookingID, "cancelled")
}
