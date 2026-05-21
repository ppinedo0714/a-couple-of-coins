package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

type importJobRepositoryIface interface {
	Create(ctx context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.ImportJob, error)
	List(ctx context.Context, userID uuid.UUID) ([]models.ImportJob, error)
}

type csvImporterIface interface {
	ProcessCSV(jobID uuid.UUID, accountID uuid.UUID, userID uuid.UUID, fileContent []byte)
}

type accountRepositoryIface interface {
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
}

type ImportsHandler struct {
	jobRepo     importJobRepositoryIface
	importer    csvImporterIface
	accountRepo accountRepositoryIface
}

func NewImportsHandler(
	jobRepo importJobRepositoryIface,
	importer csvImporterIface,
	accountRepo accountRepositoryIface,
) *ImportsHandler {
	return &ImportsHandler{
		jobRepo:     jobRepo,
		importer:    importer,
		accountRepo: accountRepo,
	}
}

func (h *ImportsHandler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	accountIDStr := r.FormValue("account_id")
	if accountIDStr == "" {
		writeError(w, http.StatusBadRequest, "account_id is required")
		return
	}
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	_, err = h.accountRepo.GetByID(r.Context(), accountID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	job, err := h.jobRepo.Create(r.Context(), userID, fileHeader.Filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create import job")
		return
	}

	go h.importer.ProcessCSV(job.ID, accountID, userID, fileBytes)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id": job.ID,
		"status": job.Status,
	})
}

func (h *ImportsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobs, err := h.jobRepo.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if jobs == nil {
		jobs = []models.ImportJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (h *ImportsHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "import job not found")
		return
	}

	job, err := h.jobRepo.GetByID(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "import job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, job)
}
