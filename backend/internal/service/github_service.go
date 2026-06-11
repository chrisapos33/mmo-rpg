package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	gh "github.com/chrisapos3/mmo-rpg/internal/github"
	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/scoring"
)

type GitHubService struct {
	ghRepo       *repository.GitHubRepo
	signalSvc    *SignalService
	clientID     string
	clientSecret string
	redirectURL  string
	frontendURL  string
	mockGitHub   bool

	// In-memory state store for OAuth CSRF protection.
	// Each entry has a 10-minute TTL; cleaned up lazily on read.
	states sync.Map // string → oauthState
}

type oauthState struct {
	userID  int64
	expires time.Time
}

func NewGitHubService(
	ghRepo *repository.GitHubRepo,
	signalSvc *SignalService,
	clientID, clientSecret, redirectURL, frontendURL string,
	mockGitHub bool,
) *GitHubService {
	return &GitHubService{
		ghRepo:       ghRepo,
		signalSvc:    signalSvc,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		frontendURL:  frontendURL,
		mockGitHub:   mockGitHub,
	}
}

// Configured returns false when GitHub OAuth credentials are absent.
func (s *GitHubService) Configured() bool {
	return s.clientID != "" && s.clientSecret != ""
}

// GetAuthorizeURL generates a CSRF state token and returns the GitHub OAuth URL.
func (s *GitHubService) GetAuthorizeURL(userID int64) (string, error) {
	if !s.Configured() {
		return "", errors.New("GitHub integration is not configured on this server")
	}
	state, err := randomHex(16)
	if err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	s.states.Store(state, oauthState{userID: userID, expires: time.Now().Add(10 * time.Minute)})
	return gh.AuthorizeURL(s.clientID, s.redirectURL, state), nil
}

// HandleCallback validates state, exchanges code, fetches GitHub data, persists connection.
// Returns the user_id on success so the handler can redirect appropriately.
// A background scoring job is enqueued; poll GET /api/github/scoring/status for completion.
func (s *GitHubService) HandleCallback(ctx context.Context, code, state string) (int64, error) {
	userID, err := s.consumeState(state)
	if err != nil {
		return 0, fmt.Errorf("invalid state: %w", err)
	}

	token, err := gh.ExchangeCode(ctx, s.clientID, s.clientSecret, code, s.redirectURL)
	if err != nil {
		return userID, fmt.Errorf("token exchange: %w", err)
	}

	user, err := gh.FetchUser(ctx, token)
	if err != nil {
		return userID, fmt.Errorf("fetching github user: %w", err)
	}

	repos, err := gh.FetchRepos(ctx, token)
	if err != nil {
		log.Printf("github_callback: repo fetch failed (non-fatal): %v", err)
		repos = nil
	}

	stats := gh.AggregateStats(user, repos)

	conn := &domain.GitHubConnection{
		UserID:            userID,
		GitHubUsername:    user.Login,
		GitHubUserID:      user.ID,
		AccessToken:       token, // TODO: encrypt at rest before going to production
		RepoCount:         user.PublicRepos,
		StarCount:         stats.TotalStars,
		Followers:         user.Followers,
		TopLanguages:      stats.TopLanguages,
		ContributionScore: stats.ContribScore,
	}
	if user.AvatarURL != "" {
		conn.AvatarURL = &user.AvatarURL
	}

	if _, err := s.ghRepo.Upsert(ctx, conn); err != nil {
		return userID, fmt.Errorf("saving github connection: %w", err)
	}

	if err := s.signalSvc.IngestGitHub(ctx, userID, user, stats); err != nil {
		log.Printf("github_callback [user:%d]: signal ingest failed (non-fatal): %v", userID, err)
	}
	log.Printf("github_callback [user:%d]: connected @%s — %d repos, %d stars",
		userID, user.Login, user.PublicRepos, stats.TotalStars)

	s.enqueueScoring(userID, token)
	return userID, nil
}

// GetConnection returns the GitHub connection for the user, or ErrNotFound.
func (s *GitHubService) GetConnection(ctx context.Context, userID int64) (*domain.GitHubConnection, error) {
	return s.ghRepo.FindByUserID(ctx, userID)
}

// Sync re-fetches GitHub data for a user who already has a token stored.
func (s *GitHubService) Sync(ctx context.Context, userID int64) (*domain.GitHubConnection, error) {
	conn, err := s.ghRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("no GitHub connection found — connect GitHub first")
	}

	user, err := gh.FetchUser(ctx, conn.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching github user: %w", err)
	}
	repos, _ := gh.FetchRepos(ctx, conn.AccessToken)
	stats := gh.AggregateStats(user, repos)

	conn.GitHubUsername = user.Login
	conn.RepoCount = user.PublicRepos
	conn.StarCount = stats.TotalStars
	conn.Followers = user.Followers
	conn.TopLanguages = stats.TopLanguages
	conn.ContributionScore = stats.ContribScore

	updated, err := s.ghRepo.Upsert(ctx, conn)
	if err != nil {
		return nil, err
	}
	if err := s.signalSvc.IngestGitHub(ctx, userID, user, stats); err != nil {
		log.Printf("github_sync [user:%d]: signal ingest failed (non-fatal): %v", userID, err)
	}

	s.enqueueScoring(userID, conn.AccessToken)
	return updated, nil
}

// ScoringStatus returns the DB row for the user's signal scores, which includes
// the scoring job lifecycle fields (status, started_at, done_at, error).
// Returns a zero-value row (ScoringStatus == nil) when no job has ever run.
func (s *GitHubService) ScoringStatus(ctx context.Context, userID int64) (*domain.UserSignalScore, error) {
	return s.signalSvc.GetScores(ctx, userID)
}

// ─── Background scoring ───────────────────────────────────────────────────────

// enqueueScoring atomically claims the DB scoring slot then starts the goroutine.
// Skips silently when a non-stale run is already in progress.
func (s *GitHubService) enqueueScoring(userID int64, token string) {
	started, err := s.signalSvc.StartScoringJob(context.Background(), userID)
	if err != nil {
		log.Printf("scoring [user:%d]: failed to start job: %v", userID, err)
		return
	}
	if !started {
		log.Printf("scoring [user:%d]: already running, skipping duplicate enqueue", userID)
		return
	}
	go s.runScoring(userID, token)
}

// runScoring executes the full ingestion→scoring pipeline in the background.
// Uses a detached context so HTTP request cancellation does not abort it.
// Panics are recovered and persisted to the DB so the server stays up.
func (s *GitHubService) runScoring(userID int64, token string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("scoring [user:%d]: panic recovered: %v", userID, r)
			if err := s.signalSvc.FailScoringJob(context.Background(), userID, fmt.Sprintf("panic: %v", r)); err != nil {
				log.Printf("scoring [user:%d]: failed to record panic: %v", userID, err)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	src := s.sourceFor(token)
	input, err := gh.Ingest(ctx, src)
	if err != nil {
		log.Printf("scoring [user:%d]: ingest failed: %v", userID, err)
		if ferr := s.signalSvc.FailScoringJob(context.Background(), userID, err.Error()); ferr != nil {
			log.Printf("scoring [user:%d]: failed to record failure: %v", userID, ferr)
		}
		return
	}

	scores := scoring.Compute(input)
	log.Printf("scoring [user:%d @%s]: output=%.0f craft=%.0f influence=%.0f collab=%.0f range=%.0f trust=%.2f",
		userID, input.Username,
		scores.Output.Percentile, scores.Craft.Percentile,
		scores.Influence.Percentile, scores.Collaboration.Percentile,
		scores.Range.Percentile, scores.Trust)

	if err := s.signalSvc.SaveGitHubScores(context.Background(), userID, input.Username, scores); err != nil {
		log.Printf("scoring [user:%d]: failed to save scores: %v", userID, err)
		if ferr := s.signalSvc.FailScoringJob(context.Background(), userID, err.Error()); ferr != nil {
			log.Printf("scoring [user:%d]: failed to record save failure: %v", userID, ferr)
		}
	}
}

// sourceFor returns the appropriate GitHubSource based on whether mock mode is enabled.
func (s *GitHubService) sourceFor(token string) gh.GitHubSource {
	if s.mockGitHub {
		return gh.NewMockGitHubSource()
	}
	return gh.NewLiveGitHubSource(token)
}

// ─── OAuth state ─────────────────────────────────────────────────────────────

// consumeState validates and removes an OAuth state from the in-memory store.
func (s *GitHubService) consumeState(state string) (int64, error) {
	v, ok := s.states.LoadAndDelete(state)
	if !ok {
		return 0, errors.New("state not found or already used")
	}
	entry := v.(oauthState)
	if time.Now().After(entry.expires) {
		return 0, errors.New("state expired")
	}
	return entry.userID, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
