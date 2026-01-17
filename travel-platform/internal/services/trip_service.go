package services

import (
	"fmt"
	"travel-platform/travel-platform/internal/models"
	"travel-platform/travel-platform/internal/repository"
)

type TripService interface {
	CreateTrip(trip *models.Trip) error
	GetTripByID(id uint) (*models.Trip, error) //iki değer döndürür//bulunan trip//hata
	GetTripByUserID(userID uint) ([]models.Trip, error)
	UpdateTrip(trip *models.Trip) error
	DeleteTrip(id uint) error
	GetPublicTrips() ([]models.Trip, error)
	SearchByDestination(destination string) ([]models.Trip, error)
}

type tripService struct { // sadece ayni paket icinden erisilebilir
	repo repository.TripRepository
}

// TripService dönüs tipi *tripService döndürüyor
// Ama TripService interface’i olarak
func NewTripService(repo repository.TripRepository) TripService { // constructor
	return &tripService{repo: repo} //& → pointer döndürür
}

func (s *tripService) CreateTrip(trip *models.Trip) error {
	// Validation
	if trip.Title == "" || trip.Destination == "" {
		return fmt.Errorf("title and destination are required")
	}
	if trip.StartDate.After(trip.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}
	return s.repo.CreateTrip(trip)
}

func (s *tripService) GetTripByID(id uint) (*models.Trip, error) {
	return s.repo.GetTripByID(id)
}

func (s *tripService) GetTripByUserID(userID uint) ([]models.Trip, error) {
	return s.repo.GetTripByUserID(userID)
}

func (s *tripService) UpdateTrip(trip *models.Trip) error {
	return s.repo.UpdateTrip(trip)
}

func (s *tripService) DeleteTrip(id uint) error {
	return s.repo.DeleteTrip(id)
}

func (s *tripService) GetPublicTrips() ([]models.Trip, error) {
	return s.repo.GetPublicTrips()
}

func (s *tripService) SearchByDestination(destination string) ([]models.Trip, error) {
	return s.repo.GetByDestination(destination)
}
