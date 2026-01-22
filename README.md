# TravelMate - Trip Planning & Social Platform

TravelMate is a comprehensive web application designed for travelers to plan their trips, interact with other travelers and receive personalized recommendations.

## 🚀 Features

- **Trip Management (CRUD):** Create, edit, view and delete trips. Add activities and expenses to keep your plans detailed and organized.
- **Social Discovery:** Explore and search for public trips shared by other users.
- **Real-time Chat:** 
    - Web-based chat via WebSockets.
    - Independent TCP Chat server and CLI client.
- **Smart Recommendations:** gRPC-based recommendation service providing personalized route and activity suggestions.
- **Profile Management:** Create and customize your traveler profile.
- **Dashboard:** A user-friendly overview of all your planned trips.

## 🛠️ Tech Stack

- **Language:** [Go (Golang)](https://golang.org/)
- **Web Framework:** [Gorilla Mux](https://github.com/gorilla/mux)
- **Database:** SQLite (using CGO-free driver)
- **ORM:** [GORM](https://gorm.io/)
- **Communication Protocols:** 
    - HTTP (Web & API)
    - WebSockets (Web Chat)
    - gRPC (Recommendation Service)
    - TCP (Chat Server)
- **Frontend:** HTML5, CSS3, Vanilla JavaScript (Template-based)

## 📂 Project Structure

- `cmd/web`: Main entry point for the web server.
- `cmd/chatclient`: CLI-based TCP chat client.
- `internal/`: Core application logic (Repository, Service, Handler, Middleware).
- `proto/`: gRPC service definitions and generated files.
- `web/`: HTML templates (`templates`) and static assets (`static`).

## ⚙️ Installation & Running

### Prerequisites
- Go 1.25 or higher installed.

### Steps

1. **Download Dependencies:**
   ```bash
   go mod download
   ```

2. **Start the Application:**
   To launch the main server (Web, TCP Chat, gRPC):
   ```bash
   go run cmd/web/main.go
   ```

3. **Access:**
   - Web Interface: `http://localhost:8080`
   - gRPC Service: `localhost:50051`
   - TCP Chat: `tcp://localhost:9090`

4. **Using the TCP Chat Client (Optional):**
   You can start the CLI client in a separate terminal:
   ```bash
   go run cmd/chatclient/main.go
   ```

## 📝 Notes
- The application automatically initializes the `travel-platform.db` SQLite database file on first run.
- gRPC and TCP services run concurrently with the main HTTP server.