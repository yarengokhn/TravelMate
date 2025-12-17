package repository

import (
	"travel-platform/travel-platform/internal/models"

	"gorm.io/gorm"
)

type TripRepository interface {
	CreateTrip(trip *models.Trip) error
	GetTripByID(id uint) (*models.Trip, error)
	GetAllTrips() ([]models.Trip, error)
	UpdateTrip(trip *models.Trip) error
	DeleteTrip(id uint) error
}

type tripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) CreateTrip(trip *models.Trip) error {
	return r.db.Create(trip).Error
}

func (r *tripRepository) GetTripByID(id uint) (*models.Trip, error) {
	var trip models.Trip
	result := r.db.First(&trip, id).Error
	if result != nil {
		return nil, result
	}
	return &trip, nil
}

func (r *tripRepository) GetAllTrips() ([]models.Trip, error) {
	var trips []models.Trip
	result := r.db.Find(&trips).Error
	if result != nil {
		return nil, result
	}
	return trips, nil
}
func (r *tripRepository) UpdateTrip(trip *models.Trip) error {
	return r.db.Save(trip).Error
}

func (r *tripRepository) DeleteTrip(id uint) error {
	return r.db.Delete(&models.Trip{}, id).Error
}
