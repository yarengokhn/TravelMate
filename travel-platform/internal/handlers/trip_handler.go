package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"travel-platform/travel-platform/internal/middleware"
	"travel-platform/travel-platform/internal/models"
	"travel-platform/travel-platform/internal/services"

	"github.com/gorilla/mux"
)

// Interface tanımı
type TripHandler interface {
	CreateTrip(w http.ResponseWriter, r *http.Request)
	GetTripByID(w http.ResponseWriter, r *http.Request)
	GetMyTrips(w http.ResponseWriter, r *http.Request)
	GetPublicTrips(w http.ResponseWriter, r *http.Request)
	SearchTrips(w http.ResponseWriter, r *http.Request)
	UpdateTrip(w http.ResponseWriter, r *http.Request)
	DeleteTrip(w http.ResponseWriter, r *http.Request)
}

// Struct (private)
type tripHandler struct {
	service services.TripService
}

// Constructor
func NewTripHandler(service services.TripService) TripHandler {
	return &tripHandler{service: service}
}

// CreateTrip - Yeni gezi oluştur (🔒 Protected)
func (h *tripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	// Context'ten userID al
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Request body'yi parse et
	var req struct {
		Title       string  `json:"title"`
		Destination string  `json:"destination"`
		StartDate   string  `json:"start_date"` // "2026-06-01"
		EndDate     string  `json:"end_date"`   // "2026-06-10"
		Description string  `json:"description"`
		Budget      float64 `json:"budget"`
		IsPublic    bool    `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// String tarihleri time.Time'a çevir
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Trip modeli oluştur
	trip := &models.Trip{
		UserID:      userID,
		Title:       req.Title,
		Destination: req.Destination,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
		Budget:      req.Budget,
		IsPublic:    req.IsPublic,
	}

	// Service'e gönder
	if err := h.service.CreateTrip(trip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Trip created successfully",
		"trip":    trip,
	})
}

// GetTripByID - ID'ye göre gezi getir
func (h *tripHandler) GetTripByID(w http.ResponseWriter, r *http.Request) {
	// URL'den ID'yi al
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	// Service'den gezi al
	trip, err := h.service.GetTripByID(uint(id))
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trip)
}

// GetMyTrips - Kullanıcının kendi gezileri (🔒 Protected)
func (h *tripHandler) GetMyTrips(w http.ResponseWriter, r *http.Request) {
	// Context'ten userID al
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Service'den gezileri al
	trips, err := h.service.GetTripByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

// GetPublicTrips - Herkese açık geziler
func (h *tripHandler) GetPublicTrips(w http.ResponseWriter, r *http.Request) {
	trips, err := h.service.GetPublicTrips()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

// SearchTrips - Destinasyona göre ara
// Örnek: /api/trips/search?destination=Paris
func (h *tripHandler) SearchTrips(w http.ResponseWriter, r *http.Request) {
	// Query parameter'ı al
	destination := r.URL.Query().Get("destination")
	if destination == "" {
		http.Error(w, "destination parameter is required", http.StatusBadRequest)
		return
	}

	// Service'den ara
	trips, err := h.service.SearchByDestination(destination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

// UpdateTrip - Gezi güncelle (🔒 Protected + Ownership kontrolü)
func (h *tripHandler) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	// Context'ten userID al
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// URL'den trip ID'yi al
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	// Gezinin sahibi mi kontrol et
	trip, err := h.service.GetTripByID(uint(id))
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	// Ownership kontrolü
	if trip.UserID != userID {
		http.Error(w, "Forbidden - You can only update your own trips", http.StatusForbidden)
		return
	}

	// Request body'yi parse et
	var req struct {
		Title       string  `json:"title"`
		Destination string  `json:"destination"`
		StartDate   string  `json:"start_date"`
		EndDate     string  `json:"end_date"`
		Description string  `json:"description"`
		Budget      float64 `json:"budget"`
		IsPublic    bool    `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trip'i güncelle
	trip.Title = req.Title
	trip.Destination = req.Destination
	trip.Description = req.Description
	trip.Budget = req.Budget
	trip.IsPublic = req.IsPublic

	// Tarihleri güncelle (eğer gönderilmişse)
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			trip.StartDate = startDate
		}
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			trip.EndDate = endDate
		}
	}

	// Service'e gönder
	if err := h.service.UpdateTrip(trip); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Trip updated successfully",
		"trip":    trip,
	})
}

// DeleteTrip - Gezi sil (🔒 Protected + Ownership kontrolü)
func (h *tripHandler) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	// Context'ten userID al
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// URL'den trip ID'yi al
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	// Gezinin sahibi mi kontrol et
	trip, err := h.service.GetTripByID(uint(id))
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	// Ownership kontrolü
	if trip.UserID != userID {
		http.Error(w, "Forbidden - You can only delete your own trips", http.StatusForbidden)
		return
	}

	// Service'den sil
	if err := h.service.DeleteTrip(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Trip deleted successfully",
	})
}
