package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"travel-platform/travel-platform/internal/database"
)

func main() {

	database.ConnectDatabase()
	r := mux.NewRouter() // Yeni bir router oluştur

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "TravelMate e hos geldiniz!")
	}).Methods("GET")

	// Sunucuyu başlat
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	// // Veritabanına bağlan
	// err := database.ConnectDatabase()
	// if err != nil {
	// 	log.Fatal("Database connection failed:", err)
	// }
	// db := database.GetDatabase()

	// userRepo := repository.NewTripRepository(db)
	// _ = userRepo // Kullanılmayan değişken hatasını önlemek için

	// log.Println("Application started successfully 🚀")
}
