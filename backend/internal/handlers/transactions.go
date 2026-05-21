package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services"
)

type transactionServiceIface interface {
	List(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]models.Transaction, int, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error)
	Update(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ClassifyUnclassified(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error)
}

type TransactionsHandler struct {
	svc transactionServiceIface
}

func NewTransactionsHandler(svc services.TransactionService) *TransactionsHandler {
	return &TransactionsHandler{svc: svc}
}

func (h *TransactionsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filters := repository.TransactionFilters{}

	if v := r.URL.Query().Get("account_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid account_id")
			return
		}
		filters.AccountID = &id
	}

	if v := r.URL.Query().Get("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		filters.CategoryID = &id
	}

	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from date, use YYYY-MM-DD")
			return
		}
		filters.From = &t
	}

	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to date, use YYYY-MM-DD")
			return
		}
		filters.To = &t
	}

	if v := r.URL.Query().Get("search"); v != "" {
		filters.Search = &v
	}

	if v := r.URL.Query().Get("unclassified"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid unclassified value")
			return
		}
		filters.Unclassified = &b
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > 200 {
			n = 200
		}
		limit = n
	}
	filters.Limit = limit

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = n
	}
	filters.Offset = offset

	txns, total, err := h.svc.List(r.Context(), userID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if txns == nil {
		txns = []models.Transaction{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": txns,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *TransactionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.svc.Create(r.Context(), userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, txn)
}

func (h *TransactionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	txn, err := h.svc.Get(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

func (h *TransactionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	// Use a raw map to detect explicit null vs absent for category_id.
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req := models.UpdateTransactionRequest{}

	if rawCat, present := raw["category_id"]; present {
		if string(rawCat) == "null" {
			req.ClearCategory = true
		} else {
			var catID uuid.UUID
			if err := json.Unmarshal(rawCat, &catID); err != nil {
				writeError(w, http.StatusBadRequest, "invalid category_id")
				return
			}
			req.CategoryID = &catID
		}
	}

	if rawDesc, present := raw["description"]; present {
		var desc string
		if err := json.Unmarshal(rawDesc, &desc); err != nil {
			writeError(w, http.StatusBadRequest, "invalid description")
			return
		}
		req.Description = &desc
	}

	if rawAmt, present := raw["amount"]; present {
		var amt float64
		if err := json.Unmarshal(rawAmt, &amt); err != nil {
			writeError(w, http.StatusBadRequest, "invalid amount")
			return
		}
		req.Amount = &amt
	}

	if rawDate, present := raw["date"]; present {
		var dateStr string
		if err := json.Unmarshal(rawDate, &dateStr); err != nil {
			writeError(w, http.StatusBadRequest, "invalid date")
			return
		}
		req.Date = &dateStr
	}

	txn, err := h.svc.Update(r.Context(), id, userID, req)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

func (h *TransactionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	err = h.svc.Delete(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionsHandler) Classify(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.svc.ClassifyUnclassified(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
