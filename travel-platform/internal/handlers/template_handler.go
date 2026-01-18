package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"travel-platform/travel-platform/internal/middleware"
	"travel-platform/travel-platform/internal/services"

	"log"
	"os"

	"github.com/gorilla/mux"
)

type TemplateHandler struct {
	templates   *template.Template
	userService services.UserService
	tripService services.TripService
}

// TemplateData - Tüm template'lere geçilecek ortak veri yapısı
type TemplateData struct {
	Title           string
	User            interface{}
	Data            interface{}
	Error           string
	Success         string
	IsAuthenticated bool
}

func NewTemplateHandler(userService services.UserService, tripService services.TripService) *TemplateHandler {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"daysBetween": func(start, end time.Time) int {
			return int(end.Sub(start).Hours() / 24)
		},
		"substr": func(s string, start, end int) string {
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"iterate": func(count uint) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"now": func() time.Time {
			return time.Now()
		},
	}

	// Tüm template dosyalarını listele
	files := []string{
		"web/templates/layout/base.html",
		"web/templates/partials/navbar.html",
		"web/templates/partials/footer.html",
		"web/templates/pages/home.html",
		"web/templates/pages/login.html",
		"web/templates/pages/register.html",
		"web/templates/pages/dashboard.html",
		"web/templates/pages/create_trip.html",
		"web/templates/pages/trip_detail.html",
		"web/templates/pages/explore.html",
	}

	// Her dosyanın varlığını kontrol et
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Fatalf("❌ Template dosyası bulunamadı: %s", file)
		}
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Fatalf("❌ Template yükleme hatası: %v", err)
	}

	log.Println("✅ Templates yüklendi")

	return &TemplateHandler{
		templates:   tmpl,
		userService: userService,
		tripService: tripService,
	}
}

// Helper method - Template render etmek için
func (h *TemplateHandler) render(w http.ResponseWriter, tmpl string, data *TemplateData) {
	err := h.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Home Page
func (h *TemplateHandler) Home(w http.ResponseWriter, r *http.Request) {
	// Public trips'leri al
	trips, err := h.tripService.GetPublicTrips()

	data := &TemplateData{
		Title: "TravelMate - Plan Your Next Adventure",
		Data:  trips,
	}

	if err != nil {
		data.Error = "Unable to load trips"
	}

	// Check if user is authenticated
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		user, _ := h.userService.GetProfile(userID)
		data.User = user
		data.IsAuthenticated = true
	}

	h.render(w, "home.html", data)
}

// Login Page (GET)
func (h *TemplateHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Login - TravelMate",
	}
	h.render(w, "login.html", data)
}

// Register Page (GET)
func (h *TemplateHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Register - TravelMate",
	}
	h.render(w, "register.html", data)
}

// Dashboard - Kullanıcının kendi trips'leri
func (h *TemplateHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	trips, err := h.tripService.GetTripByUserID(userID)

	data := &TemplateData{
		Title:           "My Trips - TravelMate",
		User:            user,
		Data:            trips,
		IsAuthenticated: true,
	}

	if err != nil {
		data.Error = "Unable to load your trips"
	}

	h.render(w, "dashboard.html", data)
}

// Create Trip Page (GET)
func (h *TemplateHandler) CreateTripPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, _ := h.userService.GetProfile(userID)

	data := &TemplateData{
		Title:           "Create New Trip - TravelMate",
		User:            user,
		IsAuthenticated: true,
	}

	h.render(w, "create_trip.html", data)
}

// Trip Detail Page
func (h *TemplateHandler) TripDetailPage(w http.ResponseWriter, r *http.Request) {
	// Get trip ID from URL
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	// Get trip from service
	trip, err := h.tripService.GetTripByID(uint(id))
	if err != nil {
		data := &TemplateData{
			Title: "Trip Not Found - TravelMate",
			Error: "The trip you're looking for doesn't exist",
		}
		h.render(w, "tripdetail.html", data)
		return
	}

	data := &TemplateData{
		Title: trip.Title + " - TravelMate",
		Data:  trip,
	}

	// Check if user is authenticated
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		user, _ := h.userService.GetProfile(userID)
		data.User = user
		data.IsAuthenticated = true
	}

	h.render(w, "trip-detail.html", data)
}

// Explore Trips Page
func (h *TemplateHandler) ExplorePage(w http.ResponseWriter, r *http.Request) {
	trips, err := h.tripService.GetPublicTrips()

	data := &TemplateData{
		Title: "Explore Trips - TravelMate",
		Data:  trips,
	}

	if err != nil {
		data.Error = "Unable to load trips"
	}

	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		user, _ := h.userService.GetProfile(userID)
		data.User = user
		data.IsAuthenticated = true
	}

	h.render(w, "explore.html", data)
}
