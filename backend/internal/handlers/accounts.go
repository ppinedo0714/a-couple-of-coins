package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services"
)

type accountServiceIface interface {
	List(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
	Create(ctx context.Context, userID uuid.UUID, req services.CreateAccountRequest) (*models.Account, error)
	Update(ctx context.Context, id, userID uuid.UUID, req services.UpdateAccountRequest) (*models.Account, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetHistory(ctx context.Context, userID uuid.UUID, req services.HistoryRequest) ([]models.BalanceSnapshot, error)
}

type AccountsHandler struct {
	svc accountServiceIface
}

func NewAccountsHandler(svc *services.AccountService) *AccountsHandler {
	return &AccountsHandler{svc: svc}
}

func (h *AccountsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if accounts == nil {
		accounts = []models.Account{}
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *AccountsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req services.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

func (h *AccountsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	account, err := h.svc.Get(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AccountsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	var req services.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	account, err := h.svc.Update(r.Context(), id, userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AccountsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	err = h.svc.Delete(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, repository.ErrHasTransactions) {
		writeError(w, http.StatusConflict, "account has transactions")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountsHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "from and to query parameters are required")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from date, use YYYY-MM-DD")
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to date, use YYYY-MM-DD")
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "day"
	}
	if interval != "day" && interval != "week" && interval != "month" {
		writeError(w, http.StatusBadRequest, "interval must be one of: day, week, month")
		return
	}

	var accountIDs []uuid.UUID
	if raw := r.URL.Query().Get("account_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, parseErr := uuid.Parse(part)
			if parseErr != nil {
				writeError(w, http.StatusBadRequest, "invalid account_id: "+part)
				return
			}
			accountIDs = append(accountIDs, id)
		}
	}

	req := services.HistoryRequest{
		AccountIDs: accountIDs,
		From:       from,
		To:         to,
		Interval:   interval,
	}

	snapshots, svcErr := h.svc.GetHistory(r.Context(), userID, req)
	if svcErr != nil {
		if errors.Is(svcErr, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, svcErr.Error())
			return
		}
		writeError(w, http.StatusBadRequest, svcErr.Error())
		return
	}

	if snapshots == nil {
		snapshots = []models.BalanceSnapshot{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"snapshots": snapshots})
}
