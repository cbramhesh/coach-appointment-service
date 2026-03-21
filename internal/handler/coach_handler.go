package handler

import (
	"coach-appointment-service/internal/apperror"
	"coach-appointment-service/internal/dto"
	"coach-appointment-service/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type CoachHandler struct {
	availService *service.AvailabilityService
	validate     *validator.Validate
}

func NewCoachHandler(as *service.AvailabilityService) *CoachHandler {
	return &CoachHandler{
		availService: as,
		validate:     validator.New(),
	}
}

func (h *CoachHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.BadRequest("INVALID_JSON", "Invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, apperror.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.availService.CreateAvailability(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}
