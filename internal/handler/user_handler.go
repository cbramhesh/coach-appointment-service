package handler

import (
	"net/http"
	"strconv"

	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/service"

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
