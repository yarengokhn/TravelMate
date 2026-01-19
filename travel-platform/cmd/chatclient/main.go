// cmd/chatclient/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"travel-platform/internal/chat"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = "9090"
)

func main() {
	// 1. Komut satırı argümanları (opsiyonel)
	host := DEFAULT_HOST
	port := DEFAULT_PORT

	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	// 2. Client oluştur
	client := chat.NewChatClient(host, port)

	// 3. Bağlan
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Connection error: %v\n", err)
	}

	// 4. Client'ı başlat (interactive mode)
	if err := client.Start(); err != nil {
		log.Fatalf("❌ Client error: %v\n", err)
	}

	fmt.Println("👋 Goodbye!")
}
