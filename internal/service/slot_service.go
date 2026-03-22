package service

import (
	"context"
	"fmt"
	"time"

	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/dto"
	"coach-appointment-service/internal/repository"
)

type SlotService struct {
	coachRepo   *repository.CoachRepo
	availRepo   *repository.AvailabilityRepo
	bookingRepo *repository.BookingRepo
}

func NewSlotService(cr *repository.CoachRepo, ar *repository.AvailabilityRepo, br *repository.BookingRepo) *SlotService {
	return &SlotService{coachRepo: cr, availRepo: ar, bookingRepo: br}
}

func (s *SlotService) GetAvailableSlots(ctx context.Context, coachID int, dateStr string) (*dto.SlotResponse, error) {
	coach, err := s.coachRepo.GetByID(ctx, coachID)
	if err != nil {
		return nil, apperror.NotFound("COACH_NOT_FOUND",
			fmt.Sprintf("Coach with ID %d not found", coachID))
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, apperror.BadRequest("INVALID_DATE", "Date must be YYYY-MM-DD format")
	}

	loc, err := time.LoadLocation(coach.Timezone)
	if err != nil {
		return nil, apperror.Internal("INVALID_TIMEZONE", "Invalid timezone configured for coach")
	}

	dateInCoachTZ := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	dayOfWeek := int(dateInCoachTZ.Weekday())

	availabilities, err := s.availRepo.GetByCoachAndDay(ctx, coachID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	var allSlots []dto.TimeSlot
	for _, avail := range availabilities {
		startStr := avail.StartTime.Format("15:04")
		endStr := avail.EndTime.Format("15:04")
		slots, err := generateSlots(startStr, endStr, dateInCoachTZ, loc)
		if err != nil {
			return nil, err
		}
		allSlots = append(allSlots, slots...)
	}

	if len(allSlots) == 0 {
		return &dto.SlotResponse{
			CoachID: coachID, Date: dateStr, Timezone: coach.Timezone,
			Slots: []dto.TimeSlot{},
		}, nil
	}

	// 7. Get booked slots for this time range
	firstStart, _ := time.Parse(time.RFC3339, allSlots[0].StartTime)
	lastEnd, _ := time.Parse(time.RFC3339, allSlots[len(allSlots)-1].EndTime)

	bookedSet := make(map[string]bool)
	if s.bookingRepo != nil {
		bookings, err := s.bookingRepo.GetConfirmedByCoachAndRange(ctx, coachID, firstStart, lastEnd)
		if err != nil {
			return nil, err
		}
		for _, b := range bookings {
			bookedSet[b.StartTime.UTC().Format(time.RFC3339)] = true
		}
	}

	// 8. Filter out booked and past slots
	now := time.Now().UTC()
	var available []dto.TimeSlot
	for _, slot := range allSlots {
		slotStart, _ := time.Parse(time.RFC3339, slot.StartTime)
		if !bookedSet[slot.StartTime] && slotStart.After(now) {
			available = append(available, slot)
		}
	}
	if available == nil {
		available = []dto.TimeSlot{}
	}

	return &dto.SlotResponse{
		CoachID: coachID, Date: dateStr, Timezone: coach.Timezone,
		Slots: available,
	}, nil
}

func generateSlots(startStr, endStr string, date time.Time, loc *time.Location) ([]dto.TimeSlot, error) {
	start, err := parseTimeStr(startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %s", startStr)
	}
	end, err := parseTimeStr(endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %s", endStr)
	}

	cursor := time.Date(date.Year(), date.Month(), date.Day(),
		start.Hour(), start.Minute(), 0, 0, loc)
	endDT := time.Date(date.Year(), date.Month(), date.Day(),
		end.Hour(), end.Minute(), 0, 0, loc)

	var slots []dto.TimeSlot
	for cursor.Add(30*time.Minute).Before(endDT) || cursor.Add(30*time.Minute).Equal(endDT) {
		slotEnd := cursor.Add(30 * time.Minute)
		slots = append(slots, dto.TimeSlot{
			StartTime: cursor.UTC().Format(time.RFC3339),
			EndTime:   slotEnd.UTC().Format(time.RFC3339),
		})
		cursor = slotEnd
	}
	return slots, nil
}

func parseTimeStr(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		t, err = time.Parse("15:04:05", s)
	}
	return t, err
}
