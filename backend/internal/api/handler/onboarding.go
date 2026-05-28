package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/repository"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

const maxUploadSize = 10 << 20 // 10 MB

type OnboardingHandler struct {
	onboarding *service.OnboardingService
}

func NewOnboardingHandler(onboarding *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboarding: onboarding}
}

// UploadCV accepts a multipart PDF upload, stores it, and triggers async AI parsing.
func (h *OnboardingHandler) UploadCV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 10 MB)")
		return
	}

	file, header, err := r.FormFile("cv")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'cv' field in form")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reading file failed")
		return
	}

	user := middleware.UserFromContext(r.Context())

	upload, err := h.onboarding.UploadCV(r.Context(), user.ID, data, header.Filename)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, upload)
}

// CVStatus returns the most recent CV upload status for the authenticated user.
func (h *OnboardingHandler) CVStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	upload, err := h.onboarding.GetCVStatus(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no CV uploaded yet")
			return
		}
		writeError(w, http.StatusInternalServerError, "status check failed")
		return
	}

	writeJSON(w, http.StatusOK, upload)
}
