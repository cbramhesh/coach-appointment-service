package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/dto"
	"coach-appointment-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	slotService    *service.SlotService
	bookingService *service.BookingService
	validate       *validator.Validate
}

func NewUserHandler(ss *service.SlotService, bs *service.BookingService) *UserHandler {
	return &UserHandler{slotService: ss, bookingService: bs, validate: validator.New()}
}

func (h *UserHandler) GetSlots(w http.ResponseWriter, r *http.Request) {
	coachIDStr := r.URL.Query().Get("coach_id")
	date := r.URL.Query().Get("date")

	if coachIDStr == "" || date == "" {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", "coach_id and date are required"))
		return
	}

	coachID, err := strconv.Atoi(coachIDStr)
	if err != nil {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", "coach_id must be a number"))
		return
	}

	result, err := h.slotService.GetAvailableSlots(r.Context(), coachID, date)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *UserHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.BadRequest("INVALID_JSON", "Invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.bookingService.CreateBooking(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *UserHandler) GetBookings(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", "user_id is required"))
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", "user_id must be a number"))
		return
	}

	result, err := h.bookingService.GetUpcomingBookings(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *UserHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bookingID, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", "Invalid booking ID"))
		return
	}

	var req dto.CancelBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.BadRequest("INVALID_JSON", "Invalid request body"))
		return
	}

	if err := h.bookingService.CancelBooking(r.Context(), bookingID, req.UserID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Booking cancelled successfully",
	})
}
