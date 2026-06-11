package handler

import (
	"errors"
	"net/http"

	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

type GitHubHandler struct {
	gh          *service.GitHubService
	frontendURL string
}

func NewGitHubHandler(gh *service.GitHubService, frontendURL string) *GitHubHandler {
	return &GitHubHandler{gh: gh, frontendURL: frontendURL}
}

// Authorize returns the GitHub OAuth URL for the authenticated user.
func (h *GitHubHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	if !h.gh.Configured() {
		writeError(w, http.StatusServiceUnavailable, "GitHub integration is not configured on this server")
		return
	}
	user := middleware.UserFromContext(r.Context())
	url, err := h.gh.GetAuthorizeURL(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate authorization URL")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// Callback handles the GitHub OAuth redirect. No auth middleware — GitHub browser redirect.
func (h *GitHubHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Redirect(w, r, h.frontendURL+"/hub?github=error&reason=missing_params", http.StatusFound)
		return
	}

	_, err := h.gh.HandleCallback(r.Context(), code, state)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/hub?github=error&reason=callback_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, h.frontendURL+"/hub?github=connected", http.StatusFound)
}

// Status returns the GitHub connection for the authenticated user, or 404.
func (h *GitHubHandler) Status(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	conn, err := h.gh.GetConnection(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"connected": false})
			return
		}
		writeError(w, http.StatusInternalServerError, "status check failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"connected": true, "connection": conn})
}

// Sync re-fetches GitHub stats for the authenticated user.
func (h *GitHubHandler) Sync(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	conn, err := h.gh.Sync(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusPreconditionFailed, "connect GitHub first")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

// ScoringStatus returns the current state of the background scoring job.
// Poll this after connecting or syncing GitHub to know when scores are ready.
// Returns {"status":"idle"} if no scoring job has ever been started for the user.
func (h *GitHubHandler) ScoringStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	scores, err := h.gh.ScoringStatus(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch scoring status")
		return
	}
	if scores.ScoringStatus == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "idle"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     *scores.ScoringStatus,
		"started_at": scores.ScoringStartedAt,
		"done_at":    scores.ScoringDoneAt,
		"error":      scores.ScoringError,
	})
}
