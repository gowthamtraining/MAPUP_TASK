package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/handler"
	"mapup-geofence/internal/middleware"
	"mapup-geofence/internal/repository"
	"mapup-geofence/internal/service"
	"mapup-geofence/internal/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	var repo repository.Repository

	if databaseURL != "" {
		log.Printf("[Database] Connecting to PostgreSQL/PostGIS at %s", databaseURL)
		pgRepo, err := repository.NewPostgresRepository(databaseURL)
		if err != nil {
			log.Printf("[Database WARNING] PostgreSQL connection failed: %v. Falling back to In-Memory repository.", err)
			repo = repository.NewMemoryRepository()
		} else {
			log.Println("[Database] Successfully connected to PostgreSQL with PostGIS extension.")
			repo = pgRepo
		}
	} else {
		log.Println("[Database] DATABASE_URL not set. Initializing high-performance In-Memory repository.")
		repo = repository.NewMemoryRepository()
	}
	defer repo.Close()

	// Seed sample demo data if empty
	seedSampleData(repo)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize Services
	geofenceService := service.NewGeofenceService(repo)
	vehicleService := service.NewVehicleService(repo)
	alertService := service.NewAlertService(repo)
	trackerService := service.NewTrackerService(repo, hub)

	// Initialize Handlers
	geofenceHandler := handler.NewGeofenceHandler(geofenceService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	locationHandler := handler.NewLocationHandler(trackerService)
	alertHandler := handler.NewAlertHandler(alertService)
	violationHandler := handler.NewViolationHandler(alertService)
	authHandler := handler.NewAuthHandler()
	wsHandler := handler.NewWebSocketHandler(hub)

	// Router setup
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key", "X-Vehicle-ID"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API Authentication Middleware (Bonus Requirement: API Key & JWT Bearer Token support)
	r.Use(middleware.AuthMiddleware)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"mapup-geofence"}`))
	})

	// Auth Endpoints (Bonus Requirement)
	r.Post("/auth/token", authHandler.GenerateToken)
	r.Post("/auth/login", authHandler.GenerateToken)
	r.Get("/auth/verify", authHandler.VerifyToken)

	// Initialize Rate Limiter for Location Updates (Bonus Requirement)
	rateRPS := 25
	burstCap := 40
	if rpsStr := os.Getenv("RATE_LIMIT_RPS"); rpsStr != "" {
		if val, err := strconv.Atoi(rpsStr); err == nil && val > 0 {
			rateRPS = val
		}
	}
	if capStr := os.Getenv("RATE_LIMIT_BURST"); capStr != "" {
		if val, err := strconv.Atoi(capStr); err == nil && val > 0 {
			burstCap = val
		}
	}
	rateLimiter := middleware.NewRateLimiter(rateRPS, burstCap, time.Second)

	// Required Endpoints
	r.Post("/geofences", geofenceHandler.CreateGeofence)
	r.Get("/geofences", geofenceHandler.GetGeofences)

	r.Post("/vehicles", vehicleHandler.CreateVehicle)
	r.Get("/vehicles", vehicleHandler.GetVehicles)

	r.With(rateLimiter.LimitMiddleware).Post("/vehicles/location", locationHandler.UpdateLocation)
	r.Get("/vehicles/location/{vehicle_id}", locationHandler.GetVehicleLocation)

	r.Post("/alerts/configure", alertHandler.ConfigureAlert)
	r.Get("/alerts", alertHandler.GetAlerts)

	r.Get("/violations/history", violationHandler.GetViolationHistory)

	// Real-time WebSocket Endpoint
	r.Get("/ws/alerts", wsHandler.ServeWs)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Geofencing & Real-Time Alert Server running on http://localhost:%s", port)
		log.Printf("📡 WebSocket alerts endpoint available at ws://localhost:%s/ws/alerts", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdownChan
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited cleanly.")
}

// seedSampleData preloads initial realistic sample data for instant exploration
func seedSampleData(repo repository.Repository) {
	ctx := context.Background()

	// Sample Geofences (San Francisco / Downtown area)
	downtownDelivery := &domain.Geofence{
		ID:          "geo_downtown_01",
		Name:        "Downtown Delivery Zone",
		Description: "Main delivery area for downtown customers",
		Category:    domain.CategoryDeliveryZone,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
		Coordinates: []domain.Coordinate{
			{37.7749, -122.4194},
			{37.7849, -122.4194},
			{37.7849, -122.4094},
			{37.7749, -122.4094},
			{37.7749, -122.4194},
		},
	}
	_ = repo.CreateGeofence(ctx, downtownDelivery)

	restrictedHarbor := &domain.Geofence{
		ID:          "geo_restricted_02",
		Name:        "Port Security Restricted Zone",
		Description: "Strict security perimeter - no unauthorized transport permitted",
		Category:    domain.CategoryRestrictedZone,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
		Coordinates: []domain.Coordinate{
			{37.7900, -122.4000},
			{37.8000, -122.4000},
			{37.8000, -122.3900},
			{37.7900, -122.3900},
			{37.7900, -122.4000},
		},
	}
	_ = repo.CreateGeofence(ctx, restrictedHarbor)

	// Sample Vehicles
	truck1 := &domain.Vehicle{
		ID:            "veh_truck_101",
		VehicleNumber: "KA-01-AB-1234",
		DriverName:    "John Doe",
		VehicleType:   "truck",
		Phone:         "+1234567890",
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
	}
	_ = repo.CreateVehicle(ctx, truck1)

	van1 := &domain.Vehicle{
		ID:            "veh_van_102",
		VehicleNumber: "MH-02-CD-5678",
		DriverName:    "Sarah Connor",
		VehicleType:   "van",
		Phone:         "+1987654321",
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
	}
	_ = repo.CreateVehicle(ctx, van1)

	// Record initial locations for seeded vehicles
	_ = repo.RecordLocation(ctx, &domain.LocationRecord{
		VehicleID: truck1.ID,
		Latitude:  37.7700,
		Longitude: -122.4250,
		Timestamp: time.Now().UTC().Add(-10 * time.Minute),
	})

	_ = repo.RecordLocation(ctx, &domain.LocationRecord{
		VehicleID: van1.ID,
		Latitude:  37.7780,
		Longitude: -122.4150,
		Timestamp: time.Now().UTC().Add(-5 * time.Minute),
	})
	_ = repo.UpdateVehicleCurrentGeofenceIDs(ctx, van1.ID, []string{downtownDelivery.ID})

	// Sample Alert Rule (monitoring entry into Downtown Delivery Zone)
	_ = repo.CreateAlertRule(ctx, &domain.AlertRule{
		AlertID:       "alert_demo_01",
		GeofenceID:    downtownDelivery.ID,
		GeofenceName:  downtownDelivery.Name,
		VehicleID:     &truck1.ID,
		VehicleNumber: &truck1.VehicleNumber,
		EventType:     domain.EventTypeEntry,
		Status:        "active",
		CreatedAt:     time.Now().UTC().Add(-15 * time.Minute),
	})

	// Sample Initial Violation Record (entry by van into delivery zone)
	_ = repo.RecordViolation(ctx, &domain.Violation{
		ID:            "viol_demo_01",
		VehicleID:     van1.ID,
		VehicleNumber: van1.VehicleNumber,
		GeofenceID:    downtownDelivery.ID,
		GeofenceName:  downtownDelivery.Name,
		EventType:     domain.EventTypeEntry,
		Latitude:      37.7780,
		Longitude:     -122.4150,
		Timestamp:     time.Now().UTC().Add(-5 * time.Minute),
	})
}
