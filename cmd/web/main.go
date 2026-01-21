package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"travel-platform/internal/chat"
	"travel-platform/internal/database"
	grpcserver "travel-platform/internal/grpc"
	"travel-platform/internal/handlers"
	"travel-platform/internal/middleware"
	"travel-platform/internal/repository"
	"travel-platform/internal/services"
	pb "travel-platform/proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

const (
	HTTP_PORT = ":8080"
	TCP_PORT  = ":9090"
	GRPC_PORT = ":50051"
)

func main() {
	// Veritabanına bağlan
	err := database.ConnectDatabase()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	db := database.GetDatabase()

	// Repository layer
	userRepo := repository.NewUserRepository(db)
	tripRepo := repository.NewTripRepository(db)

	// Service layer
	userService := services.NewUserService(userRepo)
	tripService := services.NewTripService(tripRepo)

	// Handler layer
	userHandler := handlers.NewUserHandler(userService)
	tripHandler := handlers.NewTripHandler(tripService)
	templateHandler := handlers.NewTemplateHandler(userService, tripService)
	wsHandler := handlers.NewWebSocketHandler("localhost:9090", userService)

	// Router
	r := mux.NewRouter()

	// Middleware
	r.Use(middleware.LoggingMiddleware)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	r.HandleFunc("/chat",
		middleware.AuthMiddleware(templateHandler.ChatPage)).Methods("GET")
	r.HandleFunc("/ws/chat", wsHandler.HandleWebSocket)

	// ========== WEB ROUTES (Template Pages) ==========
	// Public pages
	// Public pages (Optional Auth ekledik ki navbar user'ı tanısın)
	r.HandleFunc("/", middleware.OptionalAuthMiddleware(templateHandler.Home)).Methods("GET")
	r.HandleFunc("/login", templateHandler.LoginPage).Methods("GET")
	r.HandleFunc("/logout", templateHandler.Logout).Methods("POST")
	r.HandleFunc("/register", templateHandler.RegisterPage).Methods("GET")
	r.HandleFunc("/explore", middleware.OptionalAuthMiddleware(templateHandler.ExplorePage)).Methods("GET")

	// Protected pages (require authentication)
	r.HandleFunc("/dashboard",
		middleware.AuthMiddleware(templateHandler.Dashboard)).Methods("GET")
	r.HandleFunc("/trips/new",
		middleware.AuthMiddleware(templateHandler.CreateTripPage)).Methods("GET")
	r.HandleFunc("/trips/{id}",
		middleware.OptionalAuthMiddleware(templateHandler.TripDetailPage)).Methods("GET")
	r.HandleFunc("/profile",
		middleware.AuthMiddleware(templateHandler.ProfilePage)).Methods("GET")

	// ========== API ROUTES (JSON) ==========
	api := r.PathPrefix("/api").Subrouter()

	// User routes
	api.HandleFunc("/users/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/users/login", userHandler.Login).Methods("POST")
	api.HandleFunc("/users/logout", userHandler.Logout).Methods("POST")
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/profile",
		middleware.AuthMiddleware(userHandler.GetProfile)).Methods("GET")
	api.HandleFunc("/users/profile",
		middleware.AuthMiddleware(userHandler.UpdateProfile)).Methods("PUT")

	// Trip routes
	api.HandleFunc("/trips",
		middleware.AuthMiddleware(tripHandler.CreateTrip)).Methods("POST")
	api.HandleFunc("/trips/{id}", tripHandler.GetTripByID).Methods("GET")
	api.HandleFunc("/trips/my",
		middleware.AuthMiddleware(tripHandler.GetMyTrips)).Methods("GET")
	api.HandleFunc("/trips/public", tripHandler.GetPublicTrips).Methods("GET")
	api.HandleFunc("/trips/search", tripHandler.SearchTrips).Methods("GET")
	api.HandleFunc("/trips/{id}",
		middleware.AuthMiddleware(tripHandler.UpdateTrip)).Methods("PUT")
	api.HandleFunc("/trips/{id}",
		middleware.AuthMiddleware(tripHandler.DeleteTrip)).Methods("DELETE")

	// Sunucuyu başlat
	// ========== TCP CHAT SERVER ==========
	chatServer := chat.NewServer(TCP_PORT)
	go func() {
		fmt.Printf("💬 TCP Chat Server starting on tcp://localhost%s\n", TCP_PORT)
		if err := chatServer.Start(); err != nil {
			log.Fatal("❌ TCP Chat server error:", err)
		}
	}()

	// ========== gRPC SERVER ==========
	go func() {
		lis, err := net.Listen("tcp", GRPC_PORT)
		if err != nil {
			log.Fatalf("❌ gRPC dinlenemedi %s: %v", GRPC_PORT, err)
		}

		grpcServer := grpc.NewServer()
		recommendationServer := grpcserver.NewRecommendationServer()
		pb.RegisterRecommendationServiceServer(grpcServer, recommendationServer)

		fmt.Printf("🚀 gRPC Server: localhost%s\n", GRPC_PORT)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ gRPC hatası: %v", err)
		}
	}()

	// ========== START HTTP SERVER ==========
	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Printf("🚀 HTTP Server: http://localhost%s\n", HTTP_PORT)
	fmt.Printf("💬 TCP Chat:    tcp://localhost%s\n", TCP_PORT)
	fmt.Printf("🌐 WebSocket:   ws://localhost%s/ws/chat\n", HTTP_PORT)
	fmt.Printf("🌐 gRPC:        http://localhost%s\n", GRPC_PORT)
	fmt.Println("╚═══════════════════════════════════════════╝")

	log.Fatal(http.ListenAndServe(HTTP_PORT, r))

}
