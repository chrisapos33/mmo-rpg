package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

var validSourceTypes = map[string]bool{
	domain.SourceBlog:      true,
	domain.SourcePortfolio: true,
	domain.SourceCommunity: true,
	domain.SourceOther:     true,
}

type EvidenceHandler struct {
	signal *service.SignalService
}

func NewEvidenceHandler(signal *service.SignalService) *EvidenceHandler {
	return &EvidenceHandler{signal: signal}
}

type submitEvidenceRequest struct {
	SourceType  string  `json:"source_type"`
	Title       string  `json:"title"`
	ArtifactURL string  `json:"artifact_url"`
	Description *string `json:"description"`
}

// Submit validates and ingests a user-submitted evidence item.
// POST /api/evidence
func (h *EvidenceHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ArtifactURL = strings.TrimSpace(req.ArtifactURL)
	req.SourceType = strings.TrimSpace(req.SourceType)

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !validSourceTypes[req.SourceType] {
		writeError(w, http.StatusBadRequest, "source_type must be blog, portfolio, community, or other")
		return
	}
	if req.ArtifactURL == "" {
		writeError(w, http.StatusBadRequest, "artifact_url is required")
		return
	}
	if _, err := url.ParseRequestURI(req.ArtifactURL); err != nil {
		writeError(w, http.StatusBadRequest, "artifact_url is not a valid URL")
		return
	}

	user := middleware.UserFromContext(r.Context())

	item := &domain.EvidenceItem{
		UserID:             user.ID,
		SourceType:         req.SourceType,
		SourceKey:          req.ArtifactURL, // URL is the natural dedup key
		Title:              req.Title,
		ArtifactURL:        &req.ArtifactURL,
		Description:        req.Description,
		VerificationStatus: domain.VerifUnverified,
	}

	saved, err := h.signal.IngestManual(r.Context(), user.ID, item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save evidence")
		return
	}

	writeJSON(w, http.StatusCreated, saved)
}

// Delete removes an evidence item that belongs to the authenticated user.
// DELETE /api/evidence/:id
func (h *EvidenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	evidenceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid evidence ID")
		return
	}

	user := middleware.UserFromContext(r.Context())
	if err := h.signal.RemoveEvidence(r.Context(), user.ID, evidenceID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "evidence item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove evidence")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
