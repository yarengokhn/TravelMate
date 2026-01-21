package grpc

import (
	"context"
	"fmt"
	"math"
	pb "travel-platform/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecommendationServer struct {
	pb.UnimplementedRecommendationServiceServer
}

func NewRecommendationServer() *RecommendationServer {
	return &RecommendationServer{}
}

func (s *RecommendationServer) GetRecommendations(
	ctx context.Context,
	req *pb.RecommendationRequest,
) (*pb.RecommendationResponse, error) {

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	recommendations := s.generateRecommendations(req)

	return &pb.RecommendationResponse{
		Recommendations: recommendations,
		Message:         fmt.Sprintf("Sizin için %d öneri buldum!", len(recommendations)),
	}, nil
}

func (s *RecommendationServer) AnalyzeBudget(
	ctx context.Context,
	req *pb.BudgetAnalysisRequest,
) (*pb.BudgetAnalysisResponse, error) {

	if req.TripId == 0 {
		return nil, status.Error(codes.InvalidArgument, "trip_id gerekli")
	}

	if req.TotalBudget <= 0 {
		return nil, status.Error(codes.InvalidArgument, "total_budget 0'dan büyük olmalı")
	}

	var totalSpent float64
	categoryTotals := make(map[string]float64)

	for _, expense := range req.Expenses {
		totalSpent += expense.Amount
		categoryTotals[expense.Category] += expense.Amount
	}

	var categoryBreakdown []*pb.CategoryAnalysis
	for category, amount := range categoryTotals {
		percentage := (amount / totalSpent) * 100
		status := s.getCategoryStatus(category, percentage)

		categoryBreakdown = append(categoryBreakdown, &pb.CategoryAnalysis{
			Category:   category,
			TotalSpent: amount,
			Percentage: percentage,
			Status:     status,
		})
	}

	warnings := s.generateWarnings(req.TotalBudget, totalSpent)
	suggestions := s.generateSuggestions(categoryTotals, totalSpent)

	return &pb.BudgetAnalysisResponse{
		TotalBudget:       req.TotalBudget,
		TotalSpent:        totalSpent,
		Remaining:         req.TotalBudget - totalSpent,
		CategoryBreakdown: categoryBreakdown,
		Warnings:          warnings,
		Suggestions:       suggestions,
	}, nil
}

func (s *RecommendationServer) generateRecommendations(req *pb.RecommendationRequest) []*pb.Recommendation {
	allDestinations := []struct {
		name        string
		description string
		budget      float64
		activities  []string
		season      string
	}{
		{
			name:        "Paris, France",
			description: "Işıklar şehri, romantik kaçamaklar için mükemmel",
			budget:      1500,
			activities:  []string{"Eiffel Kulesi", "Louvre Müzesi", "Seine Nehri Turu"},
			season:      "İlkbahar/Sonbahar",
		},
		{
			name:        "Istanbul, Turkey",
			description: "Doğu ile Batı'nın buluştuğu tarih şehri",
			budget:      900,
			activities:  []string{"Ayasofya", "Boğaz Turu", "Kapalıçarşı"},
			season:      "İlkbahar/Sonbahar",
		},
		{
			name:        "Barcelona, Spain",
			description: "Muhteşem mimarisiyle Akdeniz cenneti",
			budget:      1200,
			activities:  []string{"Sagrada Familia", "Park Güell", "Plaj Keyfi"},
			season:      "Yaz",
		},
	}

	var recommendations []*pb.Recommendation

	for _, dest := range allDestinations {
		if req.MaxBudget > 0 && dest.budget > req.MaxBudget {
			continue
		}

		matchScore := s.calculateMatchScore(dest.budget, req.MaxBudget, req.PreferredDestination, dest.name)

		recommendations = append(recommendations, &pb.Recommendation{
			Destination:         dest.name,
			Description:         dest.description,
			EstimatedBudget:     dest.budget,
			SuggestedActivities: dest.activities,
			BestSeason:          dest.season,
			MatchScore:          matchScore,
		})
	}

	return recommendations
}

func (s *RecommendationServer) calculateMatchScore(destBudget, maxBudget float64, preferredDest, actualDest string) float64 {
	score := 50.0
	if maxBudget > 0 {
		budgetDiff := math.Abs(destBudget - maxBudget)
		budgetScore := math.Max(0, 30.0-(budgetDiff/100))
		score += budgetScore
	}
	if preferredDest != "" && preferredDest == actualDest {
		score += 20.0
	}
	return math.Min(score, 100.0)
}

func (s *RecommendationServer) getCategoryStatus(category string, percentage float64) string {
	idealRanges := map[string][2]float64{
		"accommodation": {30, 40},
		"food":          {25, 35},
		"transport":     {15, 25},
		"activities":    {15, 25},
	}

	if ranges, exists := idealRanges[category]; exists {
		if percentage < ranges[0] {
			return "good"
		} else if percentage <= ranges[1] {
			return "optimal"
		} else {
			return "warning"
		}
	}
	return "good"
}

func (s *RecommendationServer) generateWarnings(totalBudget, totalSpent float64) []string {
	var warnings []string
	spentPercentage := (totalSpent / totalBudget) * 100

	if spentPercentage > 90 {
		warnings = append(warnings, "⚠️ Bütçenizin %90'ını harcadınız!")
	} else if spentPercentage > 75 {
		warnings = append(warnings, "⚠️ Bütçenizin %75'ini harcadınız.")
	}

	if totalSpent > totalBudget {
		warnings = append(warnings, "🚨 Bütçenizi aştınız!")
	}
	return warnings
}

func (s *RecommendationServer) generateSuggestions(categoryTotals map[string]float64, totalSpent float64) []string {
	var suggestions []string

	for category, amount := range categoryTotals {
		percentage := (amount / totalSpent) * 100

		if category == "food" && percentage > 35 {
			suggestions = append(suggestions, "💡 Yemek harcamaları yüksek. Yerel lezzetler deneyin.")
		}
		if category == "transport" && percentage > 30 {
			suggestions = append(suggestions, "💡 Toplu taşıma kullanmayı düşünün.")
		}
		if category == "accommodation" && percentage > 45 {
			suggestions = append(suggestions, "💡 Uygun konaklama seçeneklerine bakın.")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "✅ Harcamalarınız dengeli görünüyor!")
	}
	return suggestions
}
