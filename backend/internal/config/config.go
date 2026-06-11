package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	AnthropicKey       string
	UploadDir          string
	MockAI             bool
	MockGitHub         bool
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	FrontendURL        string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Port:               os.Getenv("PORT"),
		AnthropicKey:       os.Getenv("AI_API_KEY"),
		UploadDir:          os.Getenv("UPLOAD_DIR"),
		MockAI:             os.Getenv("MOCK_AI") == "true",
		MockGitHub:         os.Getenv("MOCK_GITHUB") == "true",
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubRedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		FrontendURL:        os.Getenv("FRONTEND_URL"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.UploadDir == "" {
		cfg.UploadDir = "./uploads"
	}
	return cfg, nil
}
