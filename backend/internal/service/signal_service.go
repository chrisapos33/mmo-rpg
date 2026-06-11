package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	gh "github.com/chrisapos3/mmo-rpg/internal/github"
	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

var urlVerifyClient = &http.Client{
	Timeout: 6 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

type SignalService struct {
	signalRepo *repository.SignalRepo
}

func NewSignalService(signalRepo *repository.SignalRepo) *SignalService {
	return &SignalService{signalRepo: signalRepo}
}

// GetScores returns the user's multi-dimensional signal scores.
// Returns a zero-value score rather than ErrNotFound — callers get a consistent shape.
func (s *SignalService) GetScores(ctx context.Context, userID int64) (*domain.UserSignalScore, error) {
	score, err := s.signalRepo.GetScores(ctx, userID)
	if err == repository.ErrNotFound {
		return &domain.UserSignalScore{UserID: userID}, nil
	}
	return score, err
}

// GetEvidence returns all evidence items for a user.
func (s *SignalService) GetEvidence(ctx context.Context, userID int64) ([]*domain.EvidenceItem, error) {
	return s.signalRepo.ListEvidence(ctx, userID)
}

// IngestManual adds a user-submitted evidence item, verifies its URL, assigns signal
// events based on source_type, and recomputes scores. Returns the saved evidence item.
func (s *SignalService) IngestManual(ctx context.Context, userID int64, item *domain.EvidenceItem) (*domain.EvidenceItem, error) {
	// URL verification — HEAD request with timeout.
	if item.ArtifactURL != nil && *item.ArtifactURL != "" {
		if verifyURL(*item.ArtifactURL) {
			item.VerificationStatus = domain.VerifURLVerified
			item.VerificationConfidence = 0.60
		} else {
			item.VerificationStatus = domain.VerifUnverified
			item.VerificationConfidence = 0.20
		}
	}

	// Use URL as source_key so the same URL can't be added twice.
	if item.ArtifactURL != nil {
		item.SourceKey = *item.ArtifactURL
	}

	saved, err := s.signalRepo.UpsertEvidence(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("upserting evidence: %w", err)
	}

	// Only award signal for verified evidence.
	if saved.VerificationStatus != domain.VerifUnverified {
		events := computeManualEvents(userID, saved.ID, saved.SourceType, saved.VerificationConfidence)
		if len(events) > 0 {
			if err := s.signalRepo.ReplaceSignalEvents(ctx, saved.ID, events); err != nil {
				log.Printf("signal [user:%d]: replace events failed: %v", userID, err)
			}
		}
	}

	return saved, nil
}

// RemoveEvidence deletes an evidence item (must belong to userID).
func (s *SignalService) RemoveEvidence(ctx context.Context, userID, evidenceID int64) error {
	return s.signalRepo.DeleteEvidence(ctx, userID, evidenceID)
}

// IngestGitHub converts GitHub stats into evidence + signal events and recomputes scores.
// Called after every GitHub connect or sync.
func (s *SignalService) IngestGitHub(ctx context.Context, userID int64, user *gh.User, stats *gh.Stats) error {
	// Build metadata snapshot
	type ghMeta struct {
		Login        string   `json:"login"`
		PublicRepos  int      `json:"public_repos"`
		Followers    int      `json:"followers"`
		TotalStars   int      `json:"total_stars"`
		OriginalRepos int     `json:"original_repos"`
		TopLanguages []string `json:"top_languages"`
	}
	meta := ghMeta{
		Login:        user.Login,
		PublicRepos:  user.PublicRepos,
		Followers:    user.Followers,
		TotalStars:   stats.TotalStars,
		OriginalRepos: stats.OriginalRepos,
		TopLanguages: stats.TopLanguages,
	}
	metaBytes, _ := json.Marshal(meta)
	metaRaw := json.RawMessage(metaBytes)

	profileURL := fmt.Sprintf("https://github.com/%s", user.Login)
	title := fmt.Sprintf("GitHub: @%s", user.Login)
	confidence := 0.85 // platform_verified via OAuth token

	evidence := &domain.EvidenceItem{
		UserID:                 userID,
		SourceType:             domain.SourceGitHub,
		SourceKey:              fmt.Sprintf("%d", user.ID),
		ArtifactURL:            &profileURL,
		Title:                  title,
		VerificationStatus:     domain.VerifPlatformVerified,
		VerificationConfidence: confidence,
		MetadataJSON:           &metaRaw,
	}

	item, err := s.signalRepo.UpsertEvidence(ctx, evidence)
	if err != nil {
		return fmt.Errorf("upserting github evidence: %w", err)
	}

	events := computeGitHubEvents(userID, item.ID, stats, confidence)
	if err := s.signalRepo.ReplaceSignalEvents(ctx, item.ID, events); err != nil {
		return fmt.Errorf("replacing signal events: %w", err)
	}

	return nil
}

// StartScoringJob atomically claims the scoring slot in the DB.
// Returns false (without error) when a non-stale run is already in progress.
func (s *SignalService) StartScoringJob(ctx context.Context, userID int64) (bool, error) {
	return s.signalRepo.StartScoringJob(ctx, userID)
}

// SaveGitHubScores persists all five dimension scores + trust and marks the job done.
func (s *SignalService) SaveGitHubScores(ctx context.Context, userID int64, username string, scores scoring.Scores) error {
	return s.signalRepo.SaveGitHubScores(ctx, userID, username, scores)
}

// FailScoringJob records the error and marks the scoring job failed.
func (s *SignalService) FailScoringJob(ctx context.Context, userID int64, reason string) error {
	return s.signalRepo.FailScoringJob(ctx, userID, reason)
}

// ─── Computation ─────────────────────────────────────────────────────────────

type dimSpec struct {
	dimension   string
	basePoints  int
	weight      float64
	explanation string
}

func computeGitHubEvents(userID int64, evidenceID int64, stats *gh.Stats, confidence float64) []*domain.SignalEvent {
	stars := stats.TotalStars
	origRepos := stats.OriginalRepos
	followers := stats.User.Followers
	langCount := len(stats.TopLanguages)

	specs := []dimSpec{
		{
			dimension:   domain.DimBuilder,
			basePoints:  clamp(stars*5+origRepos*10, 0, 200),
			weight:      1.0,
			explanation: fmt.Sprintf("%d stars + %d original repos", stars, origRepos),
		},
		{
			dimension:   domain.DimExecutor,
			basePoints:  clamp(origRepos*10, 0, 150),
			weight:      0.65,
			explanation: fmt.Sprintf("%d original repos shipped", origRepos),
		},
		{
			dimension:   domain.DimSpecialist,
			basePoints:  clamp(langCount*15+origRepos*4, 0, 150),
			weight:      0.80,
			explanation: fmt.Sprintf("%d identified languages across %d repos", langCount, origRepos),
		},
		{
			dimension:   domain.DimCollaborator,
			basePoints:  clamp(followers*3, 0, 120),
			weight:      0.50,
			explanation: fmt.Sprintf("%d GitHub followers", followers),
		},
	}

	events := make([]*domain.SignalEvent, 0, len(specs))
	for _, spec := range specs {
		final := clamp(round(float64(spec.basePoints)*spec.weight*confidence), 0, 100)
		if final == 0 {
			continue
		}
		expl := spec.explanation
		ev := &domain.SignalEvent{
			UserID:               userID,
			EvidenceItemID:       &evidenceID,
			Dimension:            spec.dimension,
			BasePoints:           spec.basePoints,
			WeightMultiplier:     spec.weight,
			ConfidenceMultiplier: confidence,
			FinalPoints:          final,
			Explanation:          &expl,
		}
		events = append(events, ev)
	}
	return events
}

// computeManualEvents maps a source_type to signal dimension specs.
func computeManualEvents(userID, evidenceID int64, sourceType string, confidence float64) []*domain.SignalEvent {
	// Each source_type awards points to specific dimensions.
	// Lower than GitHub because single items at lower verification confidence.
	var specs []dimSpec
	switch sourceType {
	case domain.SourceBlog:
		specs = []dimSpec{
			{domain.DimThinker, 55, 0.85, "published writing/analysis"},
			{domain.DimSpecialist, 25, 0.60, "domain knowledge in writing"},
		}
	case domain.SourcePortfolio:
		specs = []dimSpec{
			{domain.DimBuilder, 45, 0.90, "shipped portfolio project"},
			{domain.DimSpecialist, 30, 0.75, "demonstrated domain depth"},
		}
	case domain.SourceCommunity:
		specs = []dimSpec{
			{domain.DimCollaborator, 45, 0.75, "community contribution"},
			{domain.DimThinker, 20, 0.55, "shared knowledge publicly"},
		}
	default: // other, linkedin, manual
		specs = []dimSpec{
			{domain.DimBuilder, 20, 0.65, "additional evidence"},
			{domain.DimThinker, 15, 0.55, "additional evidence"},
		}
	}

	events := make([]*domain.SignalEvent, 0, len(specs))
	for _, spec := range specs {
		final := clamp(round(float64(spec.basePoints)*spec.weight*confidence), 0, 100)
		if final == 0 {
			continue
		}
		expl := spec.explanation
		ev := &domain.SignalEvent{
			UserID:               userID,
			EvidenceItemID:       &evidenceID,
			Dimension:            spec.dimension,
			BasePoints:           spec.basePoints,
			WeightMultiplier:     spec.weight,
			ConfidenceMultiplier: confidence,
			FinalPoints:          final,
			Explanation:          &expl,
		}
		events = append(events, ev)
	}
	return events
}

// verifyURL does a HEAD request to check whether a URL is reachable.
func verifyURL(rawURL string) bool {
	req, err := http.NewRequest(http.MethodHead, rawURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Signal-Verifier/1.0")
	resp, err := urlVerifyClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 400
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round(f float64) int {
	return int(math.Round(f))
}
