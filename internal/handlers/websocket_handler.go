// internal/handlers/websocket_handler.go
package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"travel-platform/internal/database"
	"travel-platform/internal/middleware"
	"travel-platform/internal/models"
	"travel-platform/internal/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	tcpAddress  string
	userService services.UserService
}

func NewWebSocketHandler(tcpAddress string, userService services.UserService) *WebSocketHandler {
	return &WebSocketHandler{
		tcpAddress:  tcpAddress,
		userService: userService,
	}
}

type ChatTemplateData struct {
	Title           string
	User            interface{}
	IsAuthenticated bool
}

func (h *WebSocketHandler) HandleChatPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := ChatTemplateData{
		Title:           "Chat - TravelMate",
		User:            user,
		IsAuthenticated: true,
	}

	tmpl, err := template.ParseFiles("web/templates/pages/chat.html")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Render error", http.StatusInternalServerError)
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer wsConn.Close()

	tcpConn, err := net.Dial("tcp", h.tcpAddress)
	if err != nil {
		log.Printf("TCP connection error: %v", err)
		wsConn.WriteMessage(websocket.TextMessage, []byte("Error: Could not connect to chat server"))
		return
	}
	defer tcpConn.Close()

	log.Printf("✅ WebSocket client connected, proxying to TCP %s", h.tcpAddress)

	var wg sync.WaitGroup
	done := make(chan bool, 2)

	// Goroutine 1: WebSocket -> TCP
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { done <- true }()

		for {
			messageType, message, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			if messageType != websocket.TextMessage {
				continue
			}

			msg := string(message) + "\n"
			_, err = tcpConn.Write([]byte(msg))
			if err != nil {
				log.Printf("TCP write error: %v", err)
				return
			}

			log.Printf("📤 WS->TCP: %s", string(message))
		}
	}()

	// Goroutine 2: TCP -> WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { done <- true }()

		reader := bufio.NewReader(tcpConn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("TCP read error: %v", err)
				return
			}

			err = wsConn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

			log.Printf("📥 TCP->WS: %s", message[:len(message)-1])
		}
	}()

	<-done
	time.Sleep(100 * time.Millisecond)
	log.Println("🔌 WebSocket connection closed")
}

// ⭐ YENİ: API endpoint - Geçmiş mesajları almak için
func (h *WebSocketHandler) GetRoomHistory(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		http.Error(w, "room parameter required", http.StatusBadRequest)
		return
	}

	db := database.GetDatabase()

	// Room'u bul
	var room models.ChatRoom
	if err := db.Where("name = ?", roomName).First(&room).Error; err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Mesajları al
	var messages []models.ChatMessage
	db.Where("room_id = ?", room.ID).
		Order("created_at ASC").
		Limit(50).
		Find(&messages)

	// User bilgileriyle birlikte response oluştur
	type MessageResponse struct {
		Username  string    `json:"username"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
	}

	var response []MessageResponse
	for _, msg := range messages {
		var user models.User
		db.First(&user, msg.UserID)

		response = append(response, MessageResponse{
			Username:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			Message:   msg.Message,
			Timestamp: msg.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
