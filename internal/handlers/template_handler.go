package handlers

import (
	"html/template" //XSS saldırılarını önlemek için HTML'yi güvenli bir şekilde işler.
	"net/http"
	"strconv"
	"strings"
	"time"
	"travel-platform/internal/middleware"
	"travel-platform/internal/services"

	"log"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type TemplateHandler struct {
	//Handlers call the template engine, passing templates and dynamic data.
	templates   *template.Template //Parse edilmiş tüm HTML template’lerin bellekte tutulduğu yapı
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

	// Load shared templates (layouts and partials)
	// We do NOT load pages here to avoid "content" block overwrite
	layoutFiles, err := filepath.Glob("web/templates/layout/*.html")
	if err != nil {
		log.Fatal(err)
	}
	partialFiles, err := filepath.Glob("web/templates/partials/*.html")
	if err != nil {
		log.Fatal(err)
	}

	sharedFiles := append(layoutFiles, partialFiles...)

	// Create the base template with functions and shared files
	tmpl := template.New("").Funcs(funcMap)
	if len(sharedFiles) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(sharedFiles...)
		if err != nil {
			log.Fatalf("❌ Shared templates error: %v", err)
		}
	}

	log.Println("✅ Shared templates loaded")

	return &TemplateHandler{
		templates:   tmpl,
		userService: userService,
		tripService: tripService,
	}
}

// Helper method - Request-scoped render
func (h *TemplateHandler) render(w http.ResponseWriter, tmplName string, data *TemplateData) {
	// 1. Clone shared templates
	tmpl, err := h.templates.Clone()
	if err != nil {
		http.Error(w, "Template clone error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Parse the specific page
	pagePath := filepath.Join("web", "templates", "pages", tmplName)
	if _, err := os.Stat(pagePath); os.IsNotExist(err) {
		http.Error(w, "Template not found: "+tmplName, http.StatusNotFound)
		return
	}

	_, err = tmpl.ParseFiles(pagePath)
	if err != nil {
		http.Error(w, "Page parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Execute the specific page template
	err = tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
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
		h.render(w, "trip_detail.html", data)
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

	h.render(w, "trip_detail.html", data)
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
		user, err := h.userService.GetProfile(userID)
		if err == nil {
			data.User = user
			data.IsAuthenticated = true
		}
	}

	h.render(w, "explore.html", data)
}

// Chat Page
func (h *TemplateHandler) ChatPage(w http.ResponseWriter, r *http.Request) {
	// Auth middleware should have handled this, but good to check
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

	data := &TemplateData{
		Title:           "Chat - TravelMate",
		User:            user,
		IsAuthenticated: true,
	}

	h.render(w, "chat.html", data)
}

// Logout
func (h *TemplateHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		middleware.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TemplateHandler) ProfilePage(w http.ResponseWriter, r *http.Request) {
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

	data := &TemplateData{
		Title:           "Profile - TravelMate",
		User:            user,
		IsAuthenticated: true,
	}

	h.render(w, "profile.html", data)
}
