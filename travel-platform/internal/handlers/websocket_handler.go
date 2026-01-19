// internal/handlers/websocket_handler.go
package handlers

import (
	"bufio"
	"html/template"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"travel-platform/internal/middleware"
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

// TemplateData - Chat sayfası için data
type ChatTemplateData struct {
	Title           string
	User            interface{}
	IsAuthenticated bool
}

// HandleChatPage - Chat sayfasını render et (USER BİLGİSİYLE)
func (h *WebSocketHandler) HandleChatPage(w http.ResponseWriter, r *http.Request) {
	// User ID'yi context'ten al
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// User bilgisini al
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Template data hazırla
	data := ChatTemplateData{
		Title:           "Chat - TravelMate",
		User:            user,
		IsAuthenticated: true,
	}

	// Template'i parse et ve render et
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

// HandleWebSocket - WebSocket bağlantısını TCP'ye proxy'ler
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket bağlantısını upgrade et
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer wsConn.Close()

	// TCP chat server'a bağlan
	tcpConn, err := net.Dial("tcp", h.tcpAddress)
	if err != nil {
		log.Printf("TCP connection error: %v", err)
		wsConn.WriteMessage(websocket.TextMessage, []byte("Error: Could not connect to chat server"))
		return
	}
	defer tcpConn.Close()

	log.Printf("✅ WebSocket client connected, proxying to TCP %s", h.tcpAddress)

	// WaitGroup ile iki goroutine'i takip et
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
