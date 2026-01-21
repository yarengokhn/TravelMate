// internal/chat/tcp_server.go
package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"travel-platform/internal/database"
	"travel-platform/internal/models"
)

const (
	CONN_TYPE = "tcp"
	CONN_PORT = ":9090"
)

type Server struct {
	address string
	hub     *Hub
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		hub:     GetHub(),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen(CONN_TYPE, s.address)
	if err != nil {
		return fmt.Errorf("error listening: %v", err)
	}
	defer listener.Close()

	log.Printf("🚀 TCP Chat Server listening on %s\n", s.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting: %v\n", err)
			continue
		}

		log.Printf("📞 New connection from %s\n", conn.RemoteAddr().String())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// STEP 1: Get Username
	writer.WriteString("=== Welcome to TravelMate Chat ===\n")
	writer.WriteString("Enter your username: ")
	writer.Flush()

	username, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading username: %v\n", err)
		return
	}
	username = strings.TrimSpace(username)

	if username == "" {
		writer.WriteString("❌ Invalid username\n")
		writer.Flush()
		return
	}

	// STEP 2: Get User ID
	writer.WriteString("Enter your User ID: ")
	writer.Flush()

	userIDStr, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading user ID: %v\n", err)
		return
	}
	userIDStr = strings.TrimSpace(userIDStr)

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		writer.WriteString("❌ Invalid user ID\n")
		writer.Flush()
		return
	}

	// STEP 3: List available rooms
	db := database.GetDatabase()
	var rooms []models.ChatRoom
	db.Find(&rooms)

	writer.WriteString("\n📂 Available Chat Rooms:\n")
	writer.WriteString("─────────────────────────────\n")

	if len(rooms) == 0 {
		writer.WriteString("(No rooms yet)\n")
	} else {
		for i, room := range rooms {
			count := s.hub.GetRoomCount(room.ID)
			writer.WriteString(fmt.Sprintf("%d. %s (%d users online)\n",
				i+1, room.Name, count))
		}
	}
	writer.WriteString("─────────────────────────────\n")
	writer.Flush()

	// STEP 4: Room selection
	writer.WriteString("Enter room name (or create new): ")
	writer.Flush()

	roomName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading room name: %v\n", err)
		return
	}
	roomName = strings.TrimSpace(roomName)

	// STEP 5: Find or create room
	var room models.ChatRoom
	result := db.Where("name = ?", roomName).First(&room)

	if result.Error != nil {
		room = models.ChatRoom{Name: roomName}
		if err := db.Create(&room).Error; err != nil {
			writer.WriteString(fmt.Sprintf("❌ Error creating room: %v\n", err))
			writer.Flush()
			return
		}
		writer.WriteString(fmt.Sprintf("✅ Created new room: '%s'\n", roomName))
	} else {
		writer.WriteString(fmt.Sprintf("✅ Joined room: '%s'\n", roomName))
	}
	writer.Flush()

	// STEP 6: Create client
	client := &Client{
		ID:       uint(userID),
		Username: username,
		RoomID:   room.ID,
		RoomName: room.Name,
		Conn:     conn,
	}

	s.hub.Join(client)
	defer s.hub.Leave(client)

	// STEP 7: Broadcast join message
	joinMsg := fmt.Sprintf("*** %s joined the room ***", username)
	s.broadcastMessage(room.ID, joinMsg, "System", client.ID)

	// ⭐ YENİ: STEP 8: Geçmiş mesajları yükle (SON 50 MESAJ)
	var previousMessages []models.ChatMessage
	db.Where("room_id = ?", room.ID).
		Order("created_at ASC").
		Limit(50).
		Find(&previousMessages)

	if len(previousMessages) > 0 {
		writer.WriteString("\n📜 Previous messages:\n")
		writer.WriteString("─────────────────────────────\n")

		for _, msg := range previousMessages {
			// Kullanıcı bilgisini al
			var user models.User
			db.First(&user, msg.UserID)

			senderName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
			timestamp := msg.CreatedAt.Format("15:04:05")

			writer.WriteString(fmt.Sprintf("[%s] %s: %s\n",
				timestamp, senderName, msg.Message))
		}
		writer.WriteString("─────────────────────────────\n")
		writer.Flush()
	}

	// STEP 9: Welcome message
	writer.WriteString("\n╔═══════════════════════════╗\n")
	writer.WriteString(fmt.Sprintf("║ Welcome to '%s'!\n", roomName))
	writer.WriteString("╚═══════════════════════════╝\n")
	writer.WriteString("Type your messages (STOP to exit)\n\n")
	writer.Flush()

	// STEP 10: Message loop
	for {
		netData, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from %s: %v\n", username, err)
			break
		}

		message := strings.TrimSpace(netData)

		if message == "STOP" {
			log.Printf("Client %s requested disconnect\n", username)
			break
		}

		if message == "" {
			continue
		}

		// Save to database
		dbMessage := models.ChatMessage{
			RoomID:  room.ID,
			UserID:  uint(userID),
			Message: message,
		}
		db.Create(&dbMessage)

		// Broadcast to all clients in room
		s.broadcastMessage(room.ID, message, username, client.ID)

		// Echo confirmation
		timestamp := time.Now().Format("15:04:05")
		writer.WriteString(fmt.Sprintf("[%s] Message sent\n", timestamp))
		writer.Flush()
	}

	// STEP 11: Goodbye
	leaveMsg := fmt.Sprintf("*** %s left the room ***", username)
	s.broadcastMessage(room.ID, leaveMsg, "System", client.ID)

	log.Printf("Connection closed for %s\n", username)
}

func (s *Server) broadcastMessage(roomID uint, message, sender string, senderID uint) {
	clients := s.hub.GetRoomClients(roomID)
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf("[%s] %s: %s\n", timestamp, sender, message)

	for _, client := range clients {
		// Sistem mesajları herkese, normal mesajlar kendisi hariç
		if client.ID == senderID && sender != "System" {
			continue
		}

		go func(c *Client, msg string) {
			conn, ok := c.Conn.(net.Conn)
			if !ok {
				return
			}

			writer := bufio.NewWriter(conn)
			writer.WriteString(msg)
			writer.Flush()
		}(client, formattedMsg)
	}
}
