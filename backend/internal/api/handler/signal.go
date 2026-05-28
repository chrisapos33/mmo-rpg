package handler

import (
	"net/http"

	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

type SignalHandler struct {
	signal *service.SignalService
}

func NewSignalHandler(signal *service.SignalService) *SignalHandler {
	return &SignalHandler{signal: signal}
}

// GetScores returns the user's multi-dimensional signal scores.
// Always returns a valid shape — zero scores when no signal exists yet.
func (h *SignalHandler) GetScores(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	scores, err := h.signal.GetScores(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "fetching signal failed")
		return
	}
	writeJSON(w, http.StatusOK, scores)
}

// GetEvidence returns all evidence items for the user.
func (h *SignalHandler) GetEvidence(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	items, err := h.signal.GetEvidence(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "fetching evidence failed")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
