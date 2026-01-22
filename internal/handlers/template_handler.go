package handlers

import (
	"html/template"
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
	templates   *template.Template
	userService services.UserService
	tripService services.TripService
}

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
		"addFloat": func(a, b float64) float64 {
			return a + b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"percentage": func(spent, total float64) float64 {
			if total == 0 {
				return 0
			}
			return (spent / total) * 100
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
		"upper": func(s string) string {
			return strings.ToUpper(s)
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

	layoutFiles, err := filepath.Glob("web/templates/layout/*.html")
	if err != nil {
		log.Fatal(err)
	}
	partialFiles, err := filepath.Glob("web/templates/partials/*.html")
	if err != nil {
		log.Fatal(err)
	}

	sharedFiles := append(layoutFiles, partialFiles...)

	tmpl := template.New("").Funcs(funcMap)
	if len(sharedFiles) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(sharedFiles...)
		if err != nil {
			log.Fatalf("Shared templates error: %v", err)
		}
	}

	log.Println("✅ Shared templates loaded")

	return &TemplateHandler{
		templates:   tmpl,
		userService: userService,
		tripService: tripService,
	}
}

func (h *TemplateHandler) render(w http.ResponseWriter, tmplName string, data *TemplateData) {
	tmpl, err := h.templates.Clone()
	if err != nil {
		http.Error(w, "Template clone error: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	err = tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) Home(w http.ResponseWriter, r *http.Request) {
	trips, err := h.tripService.GetPublicTrips()

	data := &TemplateData{
		Title: "TravelMate - Plan Your Next Adventure",
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

	h.render(w, "home.html", data)
}

func (h *TemplateHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Login - TravelMate",
	}
	h.render(w, "login.html", data)
}

func (h *TemplateHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Register - TravelMate",
	}
	h.render(w, "register.html", data)
}

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

func (h *TemplateHandler) TripDetailPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

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

	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		user, _ := h.userService.GetProfile(userID)
		data.User = user
		data.IsAuthenticated = true
	}

	h.render(w, "trip_detail.html", data)
}

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

func (h *TemplateHandler) ChatPage(w http.ResponseWriter, r *http.Request) {
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

func (h *TemplateHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		middleware.DeleteSession(cookie.Value)
	}

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

func (h *TemplateHandler) EditTripPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	trip, err := h.tripService.GetTripByID(uint(id))
	if err != nil {
		data := &TemplateData{
			Title: "Trip Not Found - TravelMate",
			Error: "The trip you're looking for doesn't exist",
		}
		h.render(w, "edit_trip.html", data)
		return
	}

	// Check ownership
	if trip.UserID != userID {
		http.Error(w, "Forbidden - You can only edit your own trips", http.StatusForbidden)
		return
	}

	user, _ := h.userService.GetProfile(userID)

	data := &TemplateData{
		Title:           "Edit Trip - TravelMate",
		User:            user,
		Data:            trip,
		IsAuthenticated: true,
	}

	h.render(w, "edit_trip.html", data)
}
