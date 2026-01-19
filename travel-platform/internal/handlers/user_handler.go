package handlers

//Unmarshal → JSON → Struct
//Marshal → Struct → JSON

import (
	"encoding/json"
	"net/http"
	"travel-platform/internal/middleware"
	"travel-platform/internal/services"
)

type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	// DeleteProfile(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	service services.UserService //Dependency Injection prensibi
}

func NewUserHandler(service services.UserService) UserHandler {
	return &userHandler{service: service}
}

func (h *userHandler) Register(w http.ResponseWriter, r *http.Request) {
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

	// Password'ü response'dan çıkar (güvenlik)
	user.Password = ""

	//User struct'ını JSON'a çevirip geri gönderiyor
	//Marshal → Struct → JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})

}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request) {
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
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})

}

func (h *userHandler) Logout(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})

}

func (h *userHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func (h *userHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateProfile(userID, req.FirstName, req.LastName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

func (h *userHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range users {
		users[i].Password = ""
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
