package service

import (
	"context"
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
	cvRepo    *repository.CVRepo
	aiClient  *ai.Client
	uploadDir string
}

func NewOnboardingService(cvRepo *repository.CVRepo, aiClient *ai.Client, uploadDir string) *OnboardingService {
	return &OnboardingService{cvRepo: cvRepo, aiClient: aiClient, uploadDir: uploadDir}
}

// UploadCV saves the file, records the upload, and kicks off async AI parsing.
// It returns immediately — callers should poll GetCVStatus for completion.
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

// processCV runs in a background goroutine — uses a fresh context independent of the HTTP request.
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
