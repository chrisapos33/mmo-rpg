package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrisapos3/mmo-rpg/internal/ai"
	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
)

type OnboardingService struct {
	cvRepo      *repository.CVRepo
	profileRepo *repository.ProfileRepo
	aiClient    *ai.Client
	uploadDir   string
}

func NewOnboardingService(
	cvRepo *repository.CVRepo,
	profileRepo *repository.ProfileRepo,
	aiClient *ai.Client,
	uploadDir string,
) *OnboardingService {
	return &OnboardingService{
		cvRepo:      cvRepo,
		profileRepo: profileRepo,
		aiClient:    aiClient,
		uploadDir:   uploadDir,
	}
}

// UploadCV saves the file, records the upload, and kicks off async AI parsing.
// Returns immediately — callers should poll GetCVStatus for completion.
func (s *OnboardingService) UploadCV(ctx context.Context, userID int64, fileData []byte, originalName string) (*domain.CVUpload, error) {
	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating upload dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	if ext != ".pdf" {
		return nil, fmt.Errorf("only PDF files are supported")
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	storagePath := filepath.Join(s.uploadDir, filename)

	if err := os.WriteFile(storagePath, fileData, 0o644); err != nil {
		return nil, fmt.Errorf("saving file: %w", err)
	}

	upload, err := s.cvRepo.Create(ctx, userID, storagePath, originalName)
	if err != nil {
		return nil, fmt.Errorf("creating upload record: %w", err)
	}

	go s.processCV(upload.ID, storagePath)

	return upload, nil
}

func (s *OnboardingService) GetCVStatus(ctx context.Context, userID int64) (*domain.CVUpload, error) {
	return s.cvRepo.LatestByUserID(ctx, userID)
}

// GenerateBuild reads the user's parsed CV data, calls Claude, and upserts the profile.
// This call is synchronous — it blocks until the AI responds.
func (s *OnboardingService) GenerateBuild(ctx context.Context, userID int64) (*domain.Profile, error) {
	upload, err := s.cvRepo.LatestByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no CV found — upload and process your CV first")
	}
	if upload.Status != domain.CVStatusDone {
		return nil, fmt.Errorf("CV is still being processed — try again shortly")
	}
	if upload.ExtractedData == nil {
		return nil, fmt.Errorf("CV extraction data is missing")
	}

	var cvData domain.CVData
	if err := json.Unmarshal(upload.ExtractedData, &cvData); err != nil {
		return nil, fmt.Errorf("reading CV data: %w", err)
	}

	build, err := ai.GenerateBuild(ctx, s.aiClient, &cvData)
	if err != nil {
		return nil, fmt.Errorf("build generation: %w", err)
	}

	profile, err := s.profileRepo.UpsertBuild(ctx, userID, build)
	if err != nil {
		return nil, fmt.Errorf("saving build: %w", err)
	}

	log.Printf("build_gen [user:%d]: %s / %s", userID, build.Class, build.Subclass)
	return profile, nil
}

// GetBuild returns the user's profile if a build has been generated.
func (s *OnboardingService) GetBuild(ctx context.Context, userID int64) (*domain.Profile, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile.Class == nil {
		return nil, repository.ErrNotFound
	}
	return profile, nil
}

// processCV runs in a background goroutine — uses context.Background() so the
// HTTP request cancellation does not abort the Claude call.
func (s *OnboardingService) processCV(uploadID int64, storagePath string) {
	ctx := context.Background()

	text, err := ai.ExtractTextFromPDF(storagePath)
	if err != nil {
		log.Printf("cv_parse [%d]: pdf extraction failed: %v", uploadID, err)
		_ = s.cvRepo.MarkFailed(ctx, uploadID, "Failed to read PDF text: "+err.Error())
		return
	}

	data, err := ai.ParseCV(ctx, s.aiClient, text)
	if err != nil {
		log.Printf("cv_parse [%d]: claude parse failed: %v", uploadID, err)
		_ = s.cvRepo.MarkFailed(ctx, uploadID, "AI parsing failed: "+err.Error())
		return
	}

	if err := s.cvRepo.MarkDone(ctx, uploadID, data); err != nil {
		log.Printf("cv_parse [%d]: marking done failed: %v", uploadID, err)
		return
	}

	log.Printf("cv_parse [%d]: done — %d skills, %d experiences", uploadID, len(data.Skills), len(data.Experiences))
}
