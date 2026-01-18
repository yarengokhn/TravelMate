package main

import (
	"fmt"
	"log"
	"net/http"
	"travel-platform/travel-platform/internal/database"
	"travel-platform/travel-platform/internal/handlers"
	"travel-platform/travel-platform/internal/middleware"
	"travel-platform/travel-platform/internal/repository"
	"travel-platform/travel-platform/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	// Veritabanına bağlan
	err := database.ConnectDatabase()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	db := database.GetDatabase()

	// Repository layer
	userRepo := repository.NewUserRepository(db)
	tripRepo := repository.NewTripRepository(db)

	// Service layer
	userService := services.NewUserService(userRepo)
	tripService := services.NewTripService(tripRepo)

	// Handler layer
	userHandler := handlers.NewUserHandler(userService)
	tripHandler := handlers.NewTripHandler(tripService)
	templateHandler := handlers.NewTemplateHandler(userService, tripService)

	// Router
	r := mux.NewRouter()

	// Middleware
	r.Use(middleware.LoggingMiddleware)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	// ========== WEB ROUTES (Template Pages) ==========
	// Public pages
	r.HandleFunc("/", templateHandler.Home).Methods("GET")
	r.HandleFunc("/login", templateHandler.LoginPage).Methods("GET")
	r.HandleFunc("/register", templateHandler.RegisterPage).Methods("GET")
	r.HandleFunc("/explore", templateHandler.ExplorePage).Methods("GET")

	// Protected pages (require authentication)
	r.HandleFunc("/dashboard",
		middleware.AuthMiddleware(templateHandler.Dashboard)).Methods("GET")
	r.HandleFunc("/trips/new",
		middleware.AuthMiddleware(templateHandler.CreateTripPage)).Methods("GET")
	r.HandleFunc("/trips/{id}",
		templateHandler.TripDetailPage).Methods("GET")

	// ========== API ROUTES (JSON) ==========
	api := r.PathPrefix("/api").Subrouter()

	// User routes
	api.HandleFunc("/users/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/users/login", userHandler.Login).Methods("POST")
	api.HandleFunc("/users/logout", userHandler.Logout).Methods("POST")
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/profile",
		middleware.AuthMiddleware(userHandler.GetProfile)).Methods("GET")
	api.HandleFunc("/users/profile",
		middleware.AuthMiddleware(userHandler.UpdateProfile)).Methods("PUT")

	// Trip routes
	api.HandleFunc("/trips",
		middleware.AuthMiddleware(tripHandler.CreateTrip)).Methods("POST")
	api.HandleFunc("/trips/{id}", tripHandler.GetTripByID).Methods("GET")
	api.HandleFunc("/trips/my",
		middleware.AuthMiddleware(tripHandler.GetMyTrips)).Methods("GET")
	api.HandleFunc("/trips/public", tripHandler.GetPublicTrips).Methods("GET")
	api.HandleFunc("/trips/search", tripHandler.SearchTrips).Methods("GET")
	api.HandleFunc("/trips/{id}",
		middleware.AuthMiddleware(tripHandler.UpdateTrip)).Methods("PUT")
	api.HandleFunc("/trips/{id}",
		middleware.AuthMiddleware(tripHandler.DeleteTrip)).Methods("DELETE")

	// Sunucuyu başlat
	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("📁 Template directory: web/templates/")
	fmt.Println("📁 Static files: web/static/")
	log.Fatal(http.ListenAndServe(":8080", r))
}
