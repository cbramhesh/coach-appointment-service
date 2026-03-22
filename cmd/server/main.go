package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"coach-appointment-service/internal/config"
	"coach-appointment-service/internal/handler"
	"coach-appointment-service/internal/repository"
	"coach-appointment-service/internal/router"
	"coach-appointment-service/internal/service"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to MySQL database")

	coachRepo := repository.NewCoachRepo(db)
	availRepo := repository.NewAvailabilityRepo(db)
	bookingRepo := repository.NewBookingRepo(db)

	availService := service.NewAvailabilityService(availRepo, coachRepo)
	slotService := service.NewSlotService(coachRepo, availRepo, bookingRepo)
	bookingService := service.NewBookingService(bookingRepo, coachRepo, slotService)

	coachHandler := handler.NewCoachHandler(availService)
	userHandler := handler.NewUserHandler(slotService, bookingService)

	r := router.NewRouter(coachHandler, userHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
