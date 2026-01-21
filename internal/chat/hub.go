package chat

import (
	"fmt"
	"sync"
)

type Client struct {
	ID       uint
	Username string
	RoomID   uint
	RoomName string
	Conn     interface{} // net.Conn tutacağız

}

// Hub represents a chat room
type Hub struct {
	//rooms:

	// Key → RoomID
	// Value → O odadaki client listesi
	rooms map[uint][]*Client // Room ID -> Clients
	// Tek seferde tek kisi yazsin
	mu sync.RWMutex // Read Write Mutex

}

// Singleton pattern
// herkes ayni chati sistemini  kullanabilsin
// tek yönetici gibi
var (
	globalHub *Hub
	once      sync.Once
)

// “Varsa eskiyi ver, yoksa yarat”
func GetHub() *Hub {
	once.Do(func() {
		globalHub = &Hub{
			rooms: make(map[uint][]*Client),
		}
	})
	return globalHub
}

func (h *Hub) Join(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rooms[client.RoomID] = append(h.rooms[client.RoomID], client)
	fmt.Printf("%s  joined room %d (Total: %d)\n", client.Username, client.RoomID, len(h.rooms[client.RoomID]))
}

func (h *Hub) Leave(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.rooms[client.RoomID]
	for i, c := range clients {
		if c.ID == client.ID {
			h.rooms[client.RoomID] = append(clients[:i], clients[i+1:]...)
			fmt.Printf("%s left room %d (Total: %d)\n", client.Username, client.RoomID, len(h.rooms[client.RoomID]))
			break
		}
	}
	if len(h.rooms[client.RoomID]) == 0 {
		delete(h.rooms, client.RoomID)
	}
}

// GetRoomClients - Odadaki tüm client'ları döndür
func (h *Hub) GetRoomClients(roomID uint) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Kopya döndür (thread-safe)
	clients := make([]*Client, len(h.rooms[roomID]))
	copy(clients, h.rooms[roomID])
	return clients
}

// GetRoomCount - Odadaki kişi sayısı
func (h *Hub) GetRoomCount(roomID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}
