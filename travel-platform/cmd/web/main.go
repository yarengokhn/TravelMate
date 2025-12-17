package main

import (
	"log"
	"travel-platform/travel-platform/internal/database"
	"travel-platform/travel-platform/internal/repository"
)

func main() {
	// Veritabanına bağlan
	err := database.ConnectDatabase()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	db := database.GetDatabase()

	userRepo := repository.NewTripRepository(db)
	_ = userRepo // Kullanılmayan değişken hatasını önlemek için

	log.Println("Application started successfully 🚀")
}
