package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
)

type ProfileHandler struct {
	profileRepo *repository.ProfileRepo
	signalRepo  *repository.SignalRepo
	ghRepo      *repository.GitHubRepo
}

func NewProfileHandler(
	profileRepo *repository.ProfileRepo,
	signalRepo *repository.SignalRepo,
	ghRepo *repository.GitHubRepo,
) *ProfileHandler {
	return &ProfileHandler{profileRepo: profileRepo, signalRepo: signalRepo, ghRepo: ghRepo}
}

// Publish sets the authenticated user's profile to public.
func (h *ProfileHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := h.profileRepo.Publish(r.Context(), user.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusPreconditionFailed, "complete your build before publishing")
			return
		}
		writeError(w, http.StatusInternalServerError, "publish failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"published":   true,
		"profile_url": "/p/" + strconv.FormatInt(user.ID, 10),
	})
}

// GetPublic returns the public profile for a user_id, or 404 if not found/unpublished.
func (h *ProfileHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile ID")
		return
	}

	profile, err := h.profileRepo.FindPublicByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	// Signal scores and GitHub — non-fatal if absent
	signal, _ := h.signalRepo.GetScores(r.Context(), userID)
	github, _ := h.ghRepo.FindByUserID(r.Context(), userID)

	writeJSON(w, http.StatusOK, map[string]any{
		"profile": profile,
		"signal":  signal,
		"github":  github,
	})
}
