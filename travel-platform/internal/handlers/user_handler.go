package handlers

//Unmarshal → JSON → Struct
//Marshal → Struct → JSON

import (
	"encoding/json"
	"net/http"
	"travel-platform/travel-platform/internal/middleware"
	"travel-platform/travel-platform/internal/services"
)

type UserHandler struct {
	service services.UserService //Dependency Injection prensibi
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	//Gelen JSON'ı req struct'ına dönüştürüyor
	//Unmarshal → JSON → Struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req.Email, req.Password, req.FirstName, req.LastName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//User struct'ını JSON'a çevirip geri gönderiyor
	//Marshal → Struct → JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(req.Email, req.Password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	sessionId := middleware.CreateSession(user.ID, user.Email)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true, //JavaScript ile erişilemez (XSS koruması)
		MaxAge:   3600 * 24,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})

}
