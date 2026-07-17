package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

type UsersHandler struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUsersHandler(repo repository.UserRepository, jwtSecret string) *UsersHandler {
	return &UsersHandler{repo: repo, jwtSecret: jwtSecret}
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.repo.GetByID(r.Context(), userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

type updateMeRequest struct {
	Email string `json:"email"`
}

func (h *UsersHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}

	user, err := h.repo.Update(r.Context(), userID, req.Email)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		if isDuplicateEmail(err) {
			writeError(w, http.StatusConflict, "email already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}
