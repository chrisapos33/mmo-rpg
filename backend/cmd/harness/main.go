// harness is a development CLI for testing the scoring and build pipelines
// without going through HTTP or OAuth. Run from backend/:
//
//	go run ./cmd/harness [flags]
//
// Flags:
//
//	-users   comma-separated user IDs (default: all in DB)
//	-mock-ai use stub AI response instead of real Claude (default: reads MOCK_AI env)
//	-build   run build generation (default true)
//	-scores  print scores table only, skip build generation
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"github.com/chrisapos3/mmo-rpg/internal/ai"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

func main() {
	var (
		userFlag   = flag.String("users", "", "comma-separated user IDs (default: all)")
		mockAIFlag = flag.Bool("mock-ai", os.Getenv("MOCK_AI") == "true", "use mock AI instead of Claude")
		scoresOnly = flag.Bool("scores", false, "print scores only, skip build generation")
	)
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://mmorpg:mmorpg@localhost:5433/mmorpg?sslmode=disable"
	}

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	cvRepo := repository.NewCVRepo(db)
	profileRepo := repository.NewProfileRepo(db)
	signalRepo := repository.NewSignalRepo(db)
	signalSvc := service.NewSignalService(signalRepo)
	aiClient := ai.NewClient(os.Getenv("AI_API_KEY"))
	onboardingSvc := service.NewOnboardingService(cvRepo, profileRepo, signalSvc, aiClient, "./uploads", *mockAIFlag)

	ctx := context.Background()

	userIDs := resolveUsers(ctx, db, *userFlag)
	if len(userIDs) == 0 {
		log.Fatal("no users found")
	}

	fmt.Printf("MOCK_AI=%v  users=%v\n\n", *mockAIFlag, userIDs)

	for _, uid := range userIDs {
		scores, err := signalSvc.GetScores(ctx, uid)
		if err != nil {
			fmt.Printf("user %d: scores error: %v\n", uid, err)
			continue
		}

		fmt.Printf("─── User %d (%s) ───────────────────────────────\n", uid, ptrStr(scores.GitHubUsername))
		fmt.Printf("  Scoring status : %s\n", ptrStr(scores.ScoringStatus))
		fmt.Printf("  Output         : %5.1f\n", scores.OutputPercentile)
		fmt.Printf("  Craft          : %5.1f\n", scores.CraftPercentile)
		fmt.Printf("  Influence      : %5.1f\n", scores.InfluencePercentile)
		fmt.Printf("  Collaboration  : %5.1f\n", scores.CollaborationPercentile)
		fmt.Printf("  Range          : %5.1f\n", scores.RangePercentile)
		fmt.Printf("  Trust          : %.3f\n", scores.Trust)

		if *scoresOnly {
			fmt.Println()
			continue
		}

		if scores.ScoringStatus == nil || *scores.ScoringStatus != "done" {
			fmt.Printf("  (skipping build — scoring not done)\n\n")
			continue
		}

		profile, err := onboardingSvc.GenerateBuild(ctx, uid)
		if err != nil {
			fmt.Printf("  build error: %v\n\n", err)
			continue
		}

		out, _ := json.MarshalIndent(map[string]any{
			"class":        profile.Class,
			"subclass":     profile.Subclass,
			"headline":     profile.Headline,
			"summary":      profile.Summary,
			"strengths":    profile.Strengths,
			"growth_paths": profile.GrowthPaths,
		}, "  ", "  ")
		fmt.Printf("  Build:\n  %s\n\n", string(out))
	}
}

func resolveUsers(ctx context.Context, db *sqlx.DB, flag string) []int64 {
	if flag != "" {
		var ids []int64
		for _, s := range strings.Split(flag, ",") {
			s = strings.TrimSpace(s)
			id, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				log.Fatalf("invalid user ID %q: %v", s, err)
			}
			ids = append(ids, id)
		}
		return ids
	}
	var ids []int64
	if err := db.SelectContext(ctx, &ids, "SELECT id FROM users ORDER BY id"); err != nil {
		log.Fatalf("listing users: %v", err)
	}
	return ids
}

func ptrStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
