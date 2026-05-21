package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

// --- mock import job repository ---

type mockImportJobRepo struct {
	createFn  func(ctx context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error)
	getByIDFn func(ctx context.Context, id, userID uuid.UUID) (*models.ImportJob, error)
	listFn    func(ctx context.Context, userID uuid.UUID) ([]models.ImportJob, error)
}

func (m *mockImportJobRepo) Create(ctx context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error) {
	return m.createFn(ctx, userID, fileName)
}
func (m *mockImportJobRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.ImportJob, error) {
	return m.getByIDFn(ctx, id, userID)
}
func (m *mockImportJobRepo) List(ctx context.Context, userID uuid.UUID) ([]models.ImportJob, error) {
	return m.listFn(ctx, userID)
}

// --- mock CSV importer ---

type mockCSVImporter struct {
	mu       sync.Mutex
	called   bool
	doneCh   chan struct{}
	jobID    uuid.UUID
	accountID uuid.UUID
	userID   uuid.UUID
}

func newMockImporter() *mockCSVImporter {
	return &mockCSVImporter{doneCh: make(chan struct{}, 1)}
}

func (m *mockCSVImporter) ProcessCSV(jobID uuid.UUID, accountID uuid.UUID, userID uuid.UUID, _ []byte) {
	m.mu.Lock()
	m.called = true
	m.jobID = jobID
	m.accountID = accountID
	m.userID = userID
	m.mu.Unlock()
	m.doneCh <- struct{}{}
}

// --- mock account repo for imports handler ---

type mockImportAccountRepo struct {
	getByIDFn func(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
}

func (m *mockImportAccountRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	return m.getByIDFn(ctx, id, userID)
}

// --- helpers ---

func buildImportRouter(jobRepo importJobRepositoryIface, imp csvImporterIface, accRepo accountRepositoryIface) (*chi.Mux, uuid.UUID, string) {
	userID := uuid.New()
	token, _ := auth.Issue(testJWTSecret, userID, auth.TokenTTL)

	h := NewImportsHandler(jobRepo, imp, accRepo)
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))
	r.Post("/import/csv", h.UploadCSV)
	r.Get("/import/jobs", h.ListJobs)
	r.Get("/import/jobs/{id}", h.GetJob)
	return r, userID, token
}

func makeMultipartUpload(t *testing.T, accountID uuid.UUID, csvContent []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("account_id", accountID.String())
	fw, _ := w.CreateFormFile("file", "test.csv")
	_, _ = fw.Write(csvContent)
	w.Close()
	return body, w.FormDataContentType()
}

func sampleImportJob(userID uuid.UUID) *models.ImportJob {
	rowsTotal := 5
	now := time.Now()
	return &models.ImportJob{
		ID:           uuid.New(),
		UserID:       userID,
		Status:       "done",
		SourceType:   "csv",
		FileName:     "test.csv",
		RowsTotal:    &rowsTotal,
		RowsImported: 5,
		CreatedAt:    now,
		CompletedAt:  &now,
	}
}

// --- tests ---

func TestImportCSV_ValidUpload_Returns202AndSpawnsGoroutine(t *testing.T) {
	accountID := uuid.New()

	accRepo := &mockImportAccountRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID}, nil
		},
	}

	jobRepo := &mockImportJobRepo{
		createFn: func(_ context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error) {
			return &models.ImportJob{
				ID:     uuid.New(),
				UserID: userID,
				Status: "pending",
			}, nil
		},
	}

	imp := newMockImporter()
	r, _, token := buildImportRouter(jobRepo, imp, accRepo)

	body, ct := makeMultipartUpload(t, accountID, []byte("date,description,amount\n2024-01-15,Test,-10\n"))
	req := withToken(httptest.NewRequest(http.MethodPost, "/import/csv", body), token)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("want 202, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["job_id"]; !ok {
		t.Error("want job_id in response")
	}
	if resp["status"] != "pending" {
		t.Errorf("want status=pending, got %v", resp["status"])
	}

	// Verify goroutine was spawned.
	select {
	case <-imp.doneCh:
	case <-timeoutCh(t, 2):
		t.Error("ProcessCSV goroutine not called within timeout")
	}
	imp.mu.Lock()
	called := imp.called
	imp.mu.Unlock()
	if !called {
		t.Error("want ProcessCSV called")
	}
}

func TestImportCSV_AccountNotOwnedByUser_Returns404(t *testing.T) {
	accRepo := &mockImportAccountRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return nil, repository.ErrNotFound
		},
	}

	r, _, token := buildImportRouter(&mockImportJobRepo{}, newMockImporter(), accRepo)

	body, ct := makeMultipartUpload(t, uuid.New(), []byte("date,description,amount\n"))
	req := withToken(httptest.NewRequest(http.MethodPost, "/import/csv", body), token)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestImportCSV_MissingFile_Returns400(t *testing.T) {
	accRepo := &mockImportAccountRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{}, nil
		},
	}

	r, _, token := buildImportRouter(&mockImportJobRepo{}, newMockImporter(), accRepo)

	// Only account_id field, no file.
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("account_id", uuid.New().String())
	mw.Close()

	req := withToken(httptest.NewRequest(http.MethodPost, "/import/csv", body), token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestImportListJobs_Returns200WithArray(t *testing.T) {
	jobRepo := &mockImportJobRepo{
		listFn: func(_ context.Context, userID uuid.UUID) ([]models.ImportJob, error) {
			return []models.ImportJob{*sampleImportJob(userID)}, nil
		},
	}

	r, _, token := buildImportRouter(jobRepo, newMockImporter(), &mockImportAccountRepo{})

	req := withToken(httptest.NewRequest(http.MethodGet, "/import/jobs", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	var jobs []models.ImportJob
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("want 1 job, got %d", len(jobs))
	}
}

func TestImportGetJob_Found_Returns200(t *testing.T) {
	jobRepo := &mockImportJobRepo{
		getByIDFn: func(_ context.Context, id, userID uuid.UUID) (*models.ImportJob, error) {
			return sampleImportJob(userID), nil
		},
	}

	r, _, token := buildImportRouter(jobRepo, newMockImporter(), &mockImportAccountRepo{})

	req := withToken(httptest.NewRequest(http.MethodGet, "/import/jobs/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImportGetJob_NotFound_Returns404(t *testing.T) {
	jobRepo := &mockImportJobRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.ImportJob, error) {
			return nil, repository.ErrNotFound
		},
	}

	r, _, token := buildImportRouter(jobRepo, newMockImporter(), &mockImportAccountRepo{})

	req := withToken(httptest.NewRequest(http.MethodGet, "/import/jobs/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

// timeoutCh returns a channel that receives after n seconds.
func timeoutCh(t *testing.T, seconds int) <-chan struct{} {
	t.Helper()
	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(seconds) * time.Second)
		close(ch)
	}()
	return ch
}
