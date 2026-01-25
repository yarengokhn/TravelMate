package tests

import (
	"testing"
	"time"
	"travel-platform/internal/models"
	"travel-platform/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTripRepository struct {
	mock.Mock
}

func (m *MockTripRepository) CreateTrip(trip *models.Trip) error {
	args := m.Called(trip)
	return args.Error(0)
}

func (m *MockTripRepository) GetTripByID(id uint) (*models.Trip, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetTripByUserID(userID uint) ([]models.Trip, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetPublicTrips() ([]models.Trip, error) {
	args := m.Called()
	return args.Get(0).([]models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetByDestination(destination string) ([]models.Trip, error) {
	args := m.Called(destination)
	return args.Get(0).([]models.Trip), args.Error(1)
}

func (m *MockTripRepository) UpdateTrip(trip *models.Trip) error {
	args := m.Called(trip)
	return args.Error(0)
}

func (m *MockTripRepository) DeleteTrip(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateTrip_Validation(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := services.NewTripService(mockRepo)

	t.Run("Empty Title", func(t *testing.T) {
		trip := &models.Trip{Destination: "Paris"}
		err := service.CreateTrip(trip)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title and destination are required")
	})

	t.Run("Invalid Dates", func(t *testing.T) {
		trip := &models.Trip{
			Title:       "Summer Trip",
			Destination: "Paris",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now(),
		}
		err := service.CreateTrip(trip)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start date must be before end date")
	})

	t.Run("Success", func(t *testing.T) {
		trip := &models.Trip{
			Title:       "Valid Trip",
			Destination: "Paris",
			StartDate:   time.Now(),
			EndDate:     time.Now().Add(24 * time.Hour),
		}
		mockRepo.On("CreateTrip", trip).Return(nil)
		err := service.CreateTrip(trip)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
