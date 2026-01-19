// internal/handlers/websocket_handler.go
package handlers

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Production'da daha güvenli yapın
	},
}

type WebSocketHandler struct {
	tcpAddress string // TCP chat server adresi
}

func NewWebSocketHandler(tcpAddress string) *WebSocketHandler {
	return &WebSocketHandler{
		tcpAddress: tcpAddress,
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

	log.Printf("WebSocket client connected, proxying to TCP %s", h.tcpAddress)

	// İki goroutine: WebSocket -> TCP ve TCP -> WebSocket
	done := make(chan bool)

	// Goroutine 1: WebSocket'ten oku, TCP'ye yaz
	go func() {
		defer func() { done <- true }()

		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			// TCP'ye gönder
			msg := string(message) + "\n"
			_, err = tcpConn.Write([]byte(msg))
			if err != nil {
				log.Printf("TCP write error: %v", err)
				return
			}
		}
	}()

	// Goroutine 2: TCP'den oku, WebSocket'e yaz
	go func() {
		defer func() { done <- true }()

		reader := bufio.NewReader(tcpConn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("TCP read error: %v", err)
				return
			}

			// WebSocket'e gönder
			err = wsConn.WriteMessage(websocket.TextMessage, []byte(strings.TrimSpace(message)))
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	// Herhangi bir goroutine bitene kadar bekle
	<-done
	log.Println("WebSocket connection closed")
}

// HandleChatPage - Chat sayfasını render et
func (h *WebSocketHandler) HandleChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/chat.html")
}
