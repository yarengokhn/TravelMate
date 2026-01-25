package tests

import (
	"context"
	"testing"
	"travel-platform/internal/grpc"
	"travel-platform/internal/models"
	pb "travel-platform/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reuse the MockTripRepository or create a MockTripService
type MockTripService struct {
	mock.Mock
}

func (m *MockTripService) CreateTrip(trip *models.Trip) error                 { return nil }
func (m *MockTripService) GetTripByID(id uint) (*models.Trip, error)          { return nil, nil }
func (m *MockTripService) GetTripByUserID(userID uint) ([]models.Trip, error) { return nil, nil }
func (m *MockTripService) UpdateTrip(trip *models.Trip) error                 { return nil }
func (m *MockTripService) DeleteTrip(id uint) error                           { return nil }
func (m *MockTripService) GetPublicTrips() ([]models.Trip, error) {
	args := m.Called()
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripService) SearchByDestination(dest string) ([]models.Trip, error) { return nil, nil }

func TestAnalyzeBudget(t *testing.T) {
	service := new(MockTripService)
	server := grpc.NewRecommendationServer(service)

	req := &pb.BudgetAnalysisRequest{
		TripId:      1,
		TotalBudget: 1000,
		Expenses: []*pb.Expense{
			{Amount: 400, Category: "accommodation"},
			{Amount: 200, Category: "food"},
		},
	}

	resp, err := server.AnalyzeBudget(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 1000.0, resp.TotalBudget)
	assert.Equal(t, 600.0, resp.TotalSpent)
	assert.Equal(t, 400.0, resp.Remaining)
}

func TestGetRecommendations_WithTrips(t *testing.T) {
	mockService := new(MockTripService)
	server := grpc.NewRecommendationServer(mockService)

	mockTrips := []models.Trip{
		{
			Title:       "Paris Adventure",
			Destination: "Paris",
			Budget:      1200,
			IsPublic:    true,
		},
	}

	mockService.On("GetPublicTrips").Return(mockTrips, nil)

	req := &pb.RecommendationRequest{
		UserId:    1,
		MaxBudget: 2000,
	}

	resp, err := server.GetRecommendations(context.Background(), req)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Recommendations), 1)
	assert.Equal(t, "Paris", resp.Recommendations[0].Destination)
}
