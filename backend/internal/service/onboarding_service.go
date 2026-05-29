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
	mockAI      bool
}

func NewOnboardingService(
	cvRepo *repository.CVRepo,
	profileRepo *repository.ProfileRepo,
	aiClient *ai.Client,
	uploadDir string,
	mockAI bool,
) *OnboardingService {
	return &OnboardingService{
		cvRepo:      cvRepo,
		profileRepo: profileRepo,
		aiClient:    aiClient,
		uploadDir:   uploadDir,
		mockAI:      mockAI,
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

	var build *domain.BuildData
	if s.mockAI {
		build = mockBuild()
		log.Printf("build_gen [user:%d]: MOCK MODE — skipping Claude", userID)
	} else {
		var err error
		build, err = ai.GenerateBuild(ctx, s.aiClient, &cvData)
		if err != nil {
			return nil, fmt.Errorf("build generation: %w", err)
		}
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

	if s.mockAI {
		// Simulate processing delay so the frontend polling sees a real transition.
		time.Sleep(3 * time.Second)
		data := mockCVData()
		if err := s.cvRepo.MarkDone(ctx, uploadID, data); err != nil {
			log.Printf("cv_parse [%d]: marking done failed: %v", uploadID, err)
		} else {
			log.Printf("cv_parse [%d]: MOCK MODE — done", uploadID)
		}
		return
	}

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

// ─── Mock data ────────────────────────────────────────────────────────────────

func mockCVData() *domain.CVData {
	desc1 := "Led backend architecture for a real-time data platform handling 15M events/day. Reduced p99 latency by 60% through query optimisation and connection pooling."
	desc2 := "Built internal developer tooling and REST APIs consumed by 8 product teams. Introduced contract testing, cutting integration regressions by 40%."
	deg := "BSc"
	field := "Computer Science"
	year := "2018"
	email := "alex@example.com"
	loc := "Remote"
	summary := "Full-stack engineer with 6 years building distributed systems and developer tooling. Strong bias toward operational simplicity and high-leverage infrastructure work."

	return &domain.CVData{
		FullName: "Alex Hunter",
		Email:    &email,
		Location: &loc,
		Summary:  &summary,
		Experiences: []domain.CVExperience{
			{
				Company: "Acme Systems", Title: "Senior Software Engineer",
				StartDate: "2021-03", IsCurrent: true,
				Description: &desc1,
			},
			{
				Company: "Startup Labs", Title: "Software Engineer",
				StartDate: "2019-01", EndDate: strPtr("2021-02"),
				Description: &desc2,
			},
		},
		Skills:    []string{"Go", "TypeScript", "PostgreSQL", "Redis", "Kubernetes", "gRPC", "React", "Docker"},
		Education: []domain.CVEducation{{Institution: "State University", Degree: &deg, Field: &field, Year: &year}},
		Languages: []string{"English"},
		InferredSpecializations: []string{"backend systems", "distributed architecture", "developer tooling"},
	}
}

func mockBuild() *domain.BuildData {
	return &domain.BuildData{
		Class:    "The Architect",
		Subclass: "Systems Design",
		Headline: "Builds distributed systems that scale under pressure and teams that ship with confidence.",
		Summary:  "A methodical engineer who operates at the intersection of backend architecture and developer experience. Known for designing systems that are both technically rigorous and operationally sane — the kind of infrastructure work that makes every engineer around them more effective.",
		Strengths: []string{
			"Distributed systems design at scale",
			"Developer tooling and internal platform engineering",
			"Operational simplicity and performance optimisation",
		},
		GrowthPaths: []string{
			"Staff / Principal engineering track",
			"Open-source infrastructure project ownership",
			"Technical writing and architecture documentation",
		},
	}
}

func strPtr(s string) *string { return &s }
