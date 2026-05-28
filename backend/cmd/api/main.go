package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	"github.com/chrisapos3/mmo-rpg/internal/ai"
	"github.com/chrisapos3/mmo-rpg/internal/api"
	"github.com/chrisapos3/mmo-rpg/internal/api/handler"
	"github.com/chrisapos3/mmo-rpg/internal/config"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/service"
	"github.com/chrisapos3/mmo-rpg/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("upload dir: %v", err)
	}

	db, err := sqlx.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("connected to database")

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db.DB, "."); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	userRepo := repository.NewUserRepo(db)
	cvRepo := repository.NewCVRepo(db)
	profileRepo := repository.NewProfileRepo(db)
	ghRepo := repository.NewGitHubRepo(db)
	signalRepo := repository.NewSignalRepo(db)

	aiClient := ai.NewClient(cfg.AnthropicKey)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	onboardingSvc := service.NewOnboardingService(cvRepo, profileRepo, aiClient, cfg.UploadDir)
	signalSvc := service.NewSignalService(signalRepo)
	githubSvc := service.NewGitHubService(ghRepo, signalSvc, cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubRedirectURL, cfg.FrontendURL)

	profileH := handler.NewProfileHandler(profileRepo, signalRepo, ghRepo)
	exploreH := handler.NewExploreHandler(profileRepo)

	router := api.NewRouter(authSvc, onboardingSvc, githubSvc, signalSvc, profileH, exploreH, cfg.FrontendURL)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
