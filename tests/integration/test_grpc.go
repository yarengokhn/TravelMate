// test_grpc.go - Run in root directory: go run test_grpc.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
	pb "travel-platform/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🔌 Connecting to gRPC Server...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connection successful!\n")

	client := pb.NewRecommendationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║          gRPC Client - Test Started           ║")
	fmt.Println("╚═══════════════════════════════════════════════╝\n")

	// ========== TEST 1: Travel Recommendations ==========
	fmt.Println("📍 TEST 1: Travel Recommendations")
	fmt.Println("─────────────────────────────────────────────────")

	recReq := &pb.RecommendationRequest{
		UserId:               1,
		PreferredDestination: "Rome, Italy",
		MaxBudget:            1500,
	}

	recResp, err := client.GetRecommendations(ctx, recReq)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ %s\n\n", recResp.Message)

	for i, rec := range recResp.Recommendations {
		fmt.Printf("🌍 %d. %s\n", i+1, rec.Destination)
		fmt.Printf("   💰 Budget: €%.0f\n", rec.EstimatedBudget)
		fmt.Printf("   ⭐ Score: %.0f/100\n", rec.MatchScore)
		fmt.Printf("   📝 %s\n", rec.Description)
		fmt.Printf("   🌞 Best Season: %s\n", rec.BestSeason)
		fmt.Printf("   🎯 Activities:\n")
		for _, act := range rec.SuggestedActivities {
			fmt.Printf("      - %s\n", act)
		}
		fmt.Println()
	}

	// ========== TEST 2: Budget Analysis ==========
	fmt.Println("💰 TEST 2: Budget Analysis")
	fmt.Println("─────────────────────────────────────────────────")

	budgetReq := &pb.BudgetAnalysisRequest{
		TripId:      1,
		TotalBudget: 1500,
		Expenses: []*pb.Expense{
			{Category: "accommodation", Amount: 600, Currency: "EUR"},
			{Category: "food", Amount: 400, Currency: "EUR"},
			{Category: "transport", Amount: 250, Currency: "EUR"},
			{Category: "activities", Amount: 200, Currency: "EUR"},
		},
	}

	budgetResp, err := client.AnalyzeBudget(ctx, budgetReq)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("💵 Total Budget: €%.2f\n", budgetResp.TotalBudget)
	fmt.Printf("💸 Spent:        €%.2f\n", budgetResp.TotalSpent)
	fmt.Printf("💰 Remaining:    €%.2f\n\n", budgetResp.Remaining)

	fmt.Println("📊 Category Analysis:")
	for _, cat := range budgetResp.CategoryBreakdown {
		icon := "✅"
		if cat.Status == "warning" {
			icon = "⚠️"
		} else if cat.Status == "optimal" {
			icon = "🎯"
		}
		fmt.Printf("   %s %-15s: €%-8.2f (%.1f%%)\n",
			icon, cat.Category, cat.TotalSpent, cat.Percentage)
	}

	if len(budgetResp.Warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, w := range budgetResp.Warnings {
			fmt.Printf("   %s\n", w)
		}
	}

	if len(budgetResp.Suggestions) > 0 {
		fmt.Println("\n💡 Suggestions:")
		for _, s := range budgetResp.Suggestions {
			fmt.Printf("   %s\n", s)
		}
	}

	fmt.Println("\n╔═══════════════════════════════════════════════╗")
	fmt.Println("║        ✅ Tests Completed Successfully!       ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
}
