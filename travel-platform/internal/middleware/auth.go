package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	UserID    uint
	Email     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// var bloğu ile birden fazla değişken tanımlanıyor:
// sessions: Token → Session map'i
// mu: Map'e güvenli erişim için RWMutex
// map’e güvenli eşzamanlı erişim için mutex
var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

func generateSessionID() string {
	//timestamp + random
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func CreateSession(userID uint, email string) string {
	sessionID := generateSessionID()
	mu.Lock()
	sessions[sessionID] = &Session{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mu.Unlock()
	return sessionID
}

// map’e concurrent safe şekilde erişiyoruz.
func GetSession(sessionID string) (*Session, bool) {
	mu.RLock()         //read lock (okuma kilidi)
	defer mu.RUnlock() //defer → fonksiyon sonunda çalışır
	session, exists := sessions[sessionID]
	if !exists {
		return nil, false
	}
	if time.Now().After(session.ExpiresAt) {
		mu.Lock()
		delete(sessions, sessionID)
		mu.Unlock()
		return nil, false
	}
	return session, true

}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return

		}
		session, exists := GetSession(cookie.Value)
		if !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return

		}

		ctx := context.WithValue(r.Context(), "user_id", session.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	}
}
