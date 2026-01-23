package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"travel-platform/internal/services"
	pb "travel-platform/proto"

	mux "github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RecommendationHandler struct {
	grpcClient  pb.RecommendationServiceClient
	tripService services.TripService
}

func NewRecommendationHandler(tripService services.TripService) *RecommendationHandler {
	// gRPC Server'a bağlan
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return &RecommendationHandler{
		grpcClient:  pb.NewRecommendationServiceClient(conn),
		tripService: tripService,
	}
}

// GET /api/recommendations?user_id=1&max_budget=1500&destination=Paris
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 32)
	maxBudget, _ := strconv.ParseFloat(r.URL.Query().Get("max_budget"), 64)
	destination := r.URL.Query().Get("destination")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// gRPC çağrısı
	resp, err := h.grpcClient.GetRecommendations(ctx, &pb.RecommendationRequest{
		UserId:               uint32(userID),
		PreferredDestination: destination,
		MaxBudget:            maxBudget,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/budget/analyze
func (h *RecommendationHandler) AnalyzeBudget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TripID      uint32  `json:"trip_id"`
		TotalBudget float64 `json:"total_budget"`
		Expenses    []struct {
			Category string  `json:"category"`
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"expenses"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Convert expenses to protobuf format
	var expenses []*pb.Expense
	for _, exp := range req.Expenses {
		expenses = append(expenses, &pb.Expense{
			Category: exp.Category,
			Amount:   exp.Amount,
			Currency: exp.Currency,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// gRPC çağrısı
	resp, err := h.grpcClient.AnalyzeBudget(ctx, &pb.BudgetAnalysisRequest{
		TripId:      req.TripID,
		TotalBudget: req.TotalBudget,
		Expenses:    expenses,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RecommendationHandler) AnalyzeBudgetByTripID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	trip, err := h.tripService.GetTripByID(uint(tripID))
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	// Expenses'i protobuf formatına çevir
	var expenses []*pb.Expense
	for _, exp := range trip.Expenses {
		expenses = append(expenses, &pb.Expense{
			Category: exp.Category,
			Amount:   exp.Amount,
			Currency: "EUR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// gRPC çağrısı
	resp, err := h.grpcClient.AnalyzeBudget(ctx, &pb.BudgetAnalysisRequest{
		TripId:      uint32(tripID),
		TotalBudget: trip.Budget,
		Expenses:    expenses,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
