// test_grpc.go - Ana dizinde çalıştırın: go run test_grpc.go
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
	fmt.Println("🔌 gRPC Server'a bağlanılıyor...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Bağlantı başarısız: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Bağlantı başarılı!\n")

	client := pb.NewRecommendationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║          gRPC Client - Test Başladı          ║")
	fmt.Println("╚═══════════════════════════════════════════════╝\n")

	// ========== TEST 1: Seyahat Önerileri ==========
	fmt.Println("📍 TEST 1: Seyahat Önerileri")
	fmt.Println("─────────────────────────────────────────────────")

	recReq := &pb.RecommendationRequest{
		UserId:               1,
		PreferredDestination: "",
		MaxBudget:            1500,
	}

	recResp, err := client.GetRecommendations(ctx, recReq)
	if err != nil {
		log.Fatalf("❌ Hata: %v", err)
	}

	fmt.Printf("✅ %s\n\n", recResp.Message)

	for i, rec := range recResp.Recommendations {
		fmt.Printf("🌍 %d. %s\n", i+1, rec.Destination)
		fmt.Printf("   💰 Bütçe: €%.0f\n", rec.EstimatedBudget)
		fmt.Printf("   ⭐ Puan: %.0f/100\n", rec.MatchScore)
		fmt.Printf("   📝 %s\n", rec.Description)
		fmt.Printf("   🌞 En İyi Sezon: %s\n", rec.BestSeason)
		fmt.Printf("   🎯 Aktiviteler:\n")
		for _, act := range rec.SuggestedActivities {
			fmt.Printf("      - %s\n", act)
		}
		fmt.Println()
	}

	// ========== TEST 2: Bütçe Analizi ==========
	fmt.Println("💰 TEST 2: Bütçe Analizi")
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
		log.Fatalf("❌ Hata: %v", err)
	}

	fmt.Printf("💵 Toplam Bütçe: €%.2f\n", budgetResp.TotalBudget)
	fmt.Printf("💸 Harcanan:     €%.2f\n", budgetResp.TotalSpent)
	fmt.Printf("💰 Kalan:        €%.2f\n\n", budgetResp.Remaining)

	fmt.Println("📊 Kategori Analizi:")
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
		fmt.Println("\n⚠️  Uyarılar:")
		for _, w := range budgetResp.Warnings {
			fmt.Printf("   %s\n", w)
		}
	}

	if len(budgetResp.Suggestions) > 0 {
		fmt.Println("\n💡 Öneriler:")
		for _, s := range budgetResp.Suggestions {
			fmt.Printf("   %s\n", s)
		}
	}

	fmt.Println("\n╔═══════════════════════════════════════════════╗")
	fmt.Println("║        ✅ Testler Başarıyla Tamamlandı!      ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
}
