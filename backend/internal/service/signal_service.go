package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	gh "github.com/chrisapos3/mmo-rpg/internal/github"
	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
)

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

	scores, err := s.signalRepo.RecomputeScores(ctx, userID)
	if err != nil {
		return fmt.Errorf("recomputing scores: %w", err)
	}
	log.Printf("signal [user:%d]: total=%d builder=%d executor=%d specialist=%d collaborator=%d",
		userID, scores.TotalSignal, scores.BuilderScore, scores.ExecutorScore,
		scores.SpecialistScore, scores.CollaboratorScore)
	return nil
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
