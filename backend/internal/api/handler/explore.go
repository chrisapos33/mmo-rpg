package handler

import (
	"net/http"
	"strconv"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
)

const exploreMaxLimit = 50

type ExploreHandler struct {
	profileRepo *repository.ProfileRepo
}

func NewExploreHandler(profileRepo *repository.ProfileRepo) *ExploreHandler {
	return &ExploreHandler{profileRepo: profileRepo}
}

// List returns published profile cards with optional class filter and sort.
// GET /api/explore?class=The+Architect&sort=signal&limit=20&offset=0
func (h *ExploreHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	class := q.Get("class")
	sort := q.Get("sort")
	if sort != "recent" {
		sort = "signal" // default
	}

	limit := 20
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		if l > exploreMaxLimit {
			l = exploreMaxLimit
		}
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(q.Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	entries, err := h.profileRepo.ListPublished(r.Context(), class, sort, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "explore failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entries": entries,
		"classes": domain.AllClasses,
		"limit":   limit,
		"offset":  offset,
	})
}
