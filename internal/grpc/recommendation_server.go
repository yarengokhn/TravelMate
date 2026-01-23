package grpc

import (
	"context"
	"fmt"
	"math"
	"strings"
	"travel-platform/internal/services"
	pb "travel-platform/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecommendationServer struct {
	pb.UnimplementedRecommendationServiceServer
	tripService services.TripService // 👈 Ekle
}

func NewRecommendationServer(tripService services.TripService) *RecommendationServer {
	return &RecommendationServer{
		tripService: tripService, // 👈 Ekle
	}
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
		Message:         fmt.Sprintf("Found %d recommendations for you!", len(recommendations)),
	}, nil
}

func (s *RecommendationServer) AnalyzeBudget(
	ctx context.Context,
	req *pb.BudgetAnalysisRequest,
) (*pb.BudgetAnalysisResponse, error) {

	if req.TripId == 0 {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}

	if req.TotalBudget <= 0 {
		return nil, status.Error(codes.InvalidArgument, "total_budget must be greater than 0")
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
	var recommendations []*pb.Recommendation

	// 1️⃣ VERİTABANINDAN TÜM PUBLIC TRİPLERİ AL
	allTrips, err := s.tripService.GetPublicTrips()
	if err != nil || len(allTrips) == 0 {
		// Veritabanında trip yoksa, fallback olarak statik destinasyonları kullan
		return s.generateStaticRecommendations(req)
	}

	// 2️⃣ TRİPLERİ DESTİNASYONA GÖRE GRUPLA
	destinationMap := make(map[string]*destinationInfo)

	for _, trip := range allTrips {
		dest := trip.Destination

		if _, exists := destinationMap[dest]; !exists {
			destinationMap[dest] = &destinationInfo{
				destination: dest,
				budgets:     []float64{},
				activities:  make(map[string]bool),
			}
		}

		// Bütçe ekle
		if trip.Budget > 0 {
			destinationMap[dest].budgets = append(destinationMap[dest].budgets, trip.Budget)
		}

		// Aktiviteleri ekle
		for _, activity := range trip.Activities {
			destinationMap[dest].activities[activity.Name] = true
		}
	}

	// 3️⃣ HER DESTİNASYON İÇİN ÖNERİ OLUŞTUR
	for dest, info := range destinationMap {
		// Ortalama bütçe hesapla
		avgBudget := s.calculateAverageBudget(info.budgets)

		// Bütçe filtreleme
		if req.MaxBudget > 0 && avgBudget > req.MaxBudget {
			continue
		}

		// Aktiviteleri listeye çevir
		var activityList []string
		for activity := range info.activities {
			activityList = append(activityList, activity)
			if len(activityList) >= 5 { // Max 5 aktivite göster
				break
			}
		}

		// Match score hesapla
		matchScore := s.calculateMatchScore(avgBudget, req.MaxBudget, req.PreferredDestination, dest)

		// Öneri oluştur
		recommendations = append(recommendations, &pb.Recommendation{
			Destination:         dest,
			Description:         s.generateDescription(dest, len(info.budgets)),
			EstimatedBudget:     avgBudget,
			SuggestedActivities: activityList,
			BestSeason:          s.guessBestSeason(dest),
			MatchScore:          matchScore,
		})
	}

	// Match score'a göre sırala (en yüksek önce)
	s.sortRecommendationsByScore(recommendations)

	// Eğer veritabanından yeterli öneri bulunamadıysa, statik olanları ekle
	if len(recommendations) < 3 {
		staticRecs := s.generateStaticRecommendations(req)
		recommendations = append(recommendations, staticRecs...)
	}

	return recommendations
}

// 🆕 Yardımcı struct
type destinationInfo struct {
	destination string
	budgets     []float64
	activities  map[string]bool
}

// 🆕 Ortalama bütçe hesaplama
func (s *RecommendationServer) calculateAverageBudget(budgets []float64) float64 {
	if len(budgets) == 0 {
		return 1000.0 // Default bütçe
	}

	total := 0.0
	for _, b := range budgets {
		total += b
	}
	return total / float64(len(budgets))
}

// 🆕 Dinamik açıklama oluştur
func (s *RecommendationServer) generateDescription(destination string, tripCount int) string {
	return fmt.Sprintf("Popular destination with %d trips planned by our community", tripCount)
}

// 🆕 Mevsim tahmini (şimdilik basit)
func (s *RecommendationServer) guessBestSeason(destination string) string {
	// Gelişmiş bir sistemde, geçmiş triplerin tarihlerine bakabilirsiniz
	return "All Year"
}

// 🆕 Önerileri sırala
func (s *RecommendationServer) sortRecommendationsByScore(recommendations []*pb.Recommendation) {
	// Bubble sort (basit ama yeterli)
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[i].MatchScore < recommendations[j].MatchScore {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
}

// 🆕 Statik öneriler (fallback)
func (s *RecommendationServer) generateStaticRecommendations(req *pb.RecommendationRequest) []*pb.Recommendation {
	staticDestinations := []struct {
		name        string
		description string
		budget      float64
		activities  []string
		season      string
	}{
		{
			name:        "Paris, France",
			description: "City of lights, perfect for romantic getaways",
			budget:      1500,
			activities:  []string{"Eiffel Tower", "Louvre Museum", "Seine River Cruise"},
			season:      "Spring/Fall",
		},
		{
			name:        "Istanbul, Turkey",
			description: "Historic city where East meets West",
			budget:      900,
			activities:  []string{"Hagia Sophia", "Bosphorus Cruise", "Grand Bazaar"},
			season:      "Spring/Fall",
		},
		{
			name:        "Barcelona, Spain",
			description: "Mediterranean paradise with stunning architecture",
			budget:      1200,
			activities:  []string{"Sagrada Familia", "Park Güell", "Beach Time"},
			season:      "Summer",
		},
	}

	var recommendations []*pb.Recommendation

	for _, dest := range staticDestinations {
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

	// 1️⃣ BÜTÇE UYUMU (Max 30 puan)
	if maxBudget > 0 {
		budgetDiff := math.Abs(destBudget - maxBudget)
		budgetScore := math.Max(0, 30.0-(budgetDiff/100))
		score += budgetScore
	}

	// 2️⃣ DESTİNASYON TERCİHİ (Max 20 puan)
	if preferredDest != "" {
		score += s.calculateDestinationMatch(preferredDest, actualDest)
	}

	return math.Min(score, 100.0)
}

func (s *RecommendationServer) calculateDestinationMatch(preferred, actual string) float64 {
	preferred = strings.ToLower(strings.TrimSpace(preferred))
	actual = strings.ToLower(strings.TrimSpace(actual))

	if preferred == actual {
		return 20.0
	}

	if strings.Contains(actual, preferred) || strings.Contains(preferred, actual) {
		return 15.0
	}

	actualCity := strings.Split(actual, ",")[0]
	preferredCity := strings.Split(preferred, ",")[0]

	if strings.TrimSpace(actualCity) == strings.TrimSpace(preferredCity) {
		return 18.0
	}

	if strings.Contains(actualCity, preferredCity) || strings.Contains(preferredCity, actualCity) {
		return 12.0
	}

	similarity := s.calculateSimilarity(preferred, actual)
	if similarity > 0.7 {
		return similarity * 20.0
	}

	return 0.0
}

func (s *RecommendationServer) calculateSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return 0.0
	}

	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	matchCount := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if strings.Contains(w2, w1) || strings.Contains(w1, w2) {
				matchCount++
				break
			}
		}
	}

	totalWords := math.Max(float64(len(words1)), float64(len(words2)))
	if totalWords == 0 {
		return 0.0
	}

	return float64(matchCount) / totalWords
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
		warnings = append(warnings, "⚠️ You've spent 90% of your budget!")
	} else if spentPercentage > 75 {
		warnings = append(warnings, "⚠️ You've spent 75% of your budget.")
	}

	if totalSpent > totalBudget {
		warnings = append(warnings, "🚨 You've exceeded your budget!")
	}
	return warnings
}

func (s *RecommendationServer) generateSuggestions(categoryTotals map[string]float64, totalSpent float64) []string {
	var suggestions []string

	for category, amount := range categoryTotals {
		percentage := (amount / totalSpent) * 100

		if category == "food" && percentage > 35 {
			suggestions = append(suggestions, "💡 Food expenses are high. Try local cuisine.")
		}
		if category == "transport" && percentage > 30 {
			suggestions = append(suggestions, "💡 Consider using public transportation.")
		}
		if category == "accommodation" && percentage > 45 {
			suggestions = append(suggestions, "💡 Look for more affordable accommodation options.")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "✅ Your expenses look balanced!")
	}
	return suggestions
}
