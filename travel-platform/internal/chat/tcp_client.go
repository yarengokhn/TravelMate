package chat

//Test sayfasi tcp icin
import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type ClientConfig struct {
	Host string
	Port string
	Type string
}

type ChatClient struct {
	config ClientConfig
	conn   net.Conn
}

func NewChatClient(host, port string) *ChatClient {
	return &ChatClient{
		config: ClientConfig{
			Host: host,
			Port: port,
			Type: "tcp",
		},
	}
}

func (c *ChatClient) Connect() error {

	// Connect to server
	conn, err := net.Dial(c.config.Type, c.config.Host+":"+c.config.Port)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	c.conn = conn
	fmt.Println("Connected to server")
	return nil
}

func (c *ChatClient) Start() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	defer c.conn.Close()

	go c.readFromServer()

	return c.readFromStdin()
}

func (c *ChatClient) readFromServer() {
	reader := bufio.NewReader(c.conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\n❌ Connection closed by server")
			os.Exit(0)
		}

		// Mesajı ekrana yazdır
		fmt.Print(message)
	}
}

// readFromStdin - Kullanıcıdan input al ve sunucuya gönder
func (c *ChatClient) readFromStdin() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Kullanıcıdan input oku
		text, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}

		// Sunucuya gönder
		_, err = fmt.Fprintf(c.conn, text)
		if err != nil {
			return fmt.Errorf("error sending message: %v", err)
		}

		// STOP komutu ile çık
		if strings.TrimSpace(text) == "STOP" {
			fmt.Println("👋 Exiting...")
			return nil
		}
	}
}

// SendMessage - Programatik mesaj gönderme (API için)
func (c *ChatClient) SendMessage(message string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	_, err := fmt.Fprintf(c.conn, "%s\n", message)
	return err
}

// Close - Bağlantıyı kapat
func (c *ChatClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
