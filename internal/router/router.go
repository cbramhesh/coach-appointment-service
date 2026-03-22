package router

import (
	"net/http"

	"coach-appointment-service/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(ch *handler.CoachHandler, uh *handler.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/coaches", func(r chi.Router) {
		r.Post("/availability", ch.SetAvailability)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/slots", uh.GetSlots)
		r.Post("/bookings", uh.CreateBooking)
		r.Get("/bookings", uh.GetBookings)
		r.Patch("/bookings/{id}", uh.CancelBooking)
	})

	return r
}
