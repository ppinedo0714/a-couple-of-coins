package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services"
)

type categoryServiceIface interface {
	List(ctx context.Context, userID uuid.UUID) ([]models.Category, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*models.Category, error)
	Create(ctx context.Context, userID uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error)
	Update(ctx context.Context, id, userID uuid.UUID, req services.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type CategoriesHandler struct {
	svc categoryServiceIface
}

func NewCategoriesHandler(svc *services.CategoryService) *CategoriesHandler {
	return &CategoriesHandler{svc: svc}
}

func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cats, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if cats == nil {
		cats = []models.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req services.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		if isDuplicateName(err) {
			writeError(w, http.StatusConflict, "category name already exists in this scope")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	var req services.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.Update(r.Context(), id, userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	if err != nil {
		if isDuplicateName(err) {
			writeError(w, http.StatusConflict, "category name already exists in this scope")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cat)
}

func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	err = h.svc.Delete(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isDuplicateName(err error) bool {
	if err == nil {
		return false
	}
	// pgx wraps pgconn.PgError; unique_violation = 23505
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
