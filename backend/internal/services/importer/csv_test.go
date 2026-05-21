package importer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

// --- mock job repo ---

type mockJobRepo struct {
	mu sync.Mutex

	updateStatusCalls    []statusCall
	updateRowsTotalCalls []rowsTotalCall
	incrementCalls       []incrementCall
	completeCalls        []completeCall
}

type statusCall struct {
	id     uuid.UUID
	status string
}
type rowsTotalCall struct {
	id    uuid.UUID
	total int
}
type incrementCall struct {
	id    uuid.UUID
	count int
}
type completeCall struct {
	id     uuid.UUID
	status string
}

func (m *mockJobRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateStatusCalls = append(m.updateStatusCalls, statusCall{id, status})
	return nil
}
func (m *mockJobRepo) UpdateRowsTotal(_ context.Context, id uuid.UUID, total int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateRowsTotalCalls = append(m.updateRowsTotalCalls, rowsTotalCall{id, total})
	return nil
}
func (m *mockJobRepo) IncrementRowsImported(_ context.Context, id uuid.UUID, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.incrementCalls = append(m.incrementCalls, incrementCall{id, count})
	return nil
}
func (m *mockJobRepo) Complete(_ context.Context, id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completeCalls = append(m.completeCalls, completeCall{id, status})
	return nil
}

// --- mock transaction repo ---

type mockTxRepo struct {
	mu              sync.Mutex
	bulkInsertCalls [][]repository.CreateTransactionParams
	failBulkInsert  bool
}

func (m *mockTxRepo) BulkInsert(_ context.Context, params []repository.CreateTransactionParams) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failBulkInsert {
		return 0, context.DeadlineExceeded
	}
	m.bulkInsertCalls = append(m.bulkInsertCalls, params)
	return len(params), nil
}

// --- mock account repo (implements repository.AccountRepository) ---

type mockAccRepo struct{}

func (m *mockAccRepo) List(_ context.Context, _ uuid.UUID) ([]models.Account, error) { return nil, nil }
func (m *mockAccRepo) GetByID(_ context.Context, id, _ uuid.UUID) (*models.Account, error) {
	return &models.Account{ID: id, Balance: 0}, nil
}
func (m *mockAccRepo) Create(_ context.Context, _ uuid.UUID, _, _, _ string, _ float64) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepo) Update(_ context.Context, _, _ uuid.UUID, _ repository.AccountUpdateFields) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepo) Delete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockAccRepo) UpdateBalance(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ float64) error {
	return nil
}
func (m *mockAccRepo) UpdateBalanceDirect(_ context.Context, _ uuid.UUID, _ float64) error {
	return nil
}
func (m *mockAccRepo) ListBalanceSnapshots(_ context.Context, _ []uuid.UUID, _, _ time.Time, _ string) ([]models.BalanceSnapshot, error) {
	return nil, nil
}
func (m *mockAccRepo) UpsertBalanceSnapshot(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ time.Time, _ float64) error {
	return nil
}
func (m *mockAccRepo) UpsertBalanceSnapshotDirect(_ context.Context, _ uuid.UUID, _ time.Time, _ float64) error {
	return nil
}

// Compile-time check that mockAccRepo implements repository.AccountRepository.
var _ repository.AccountRepository = (*mockAccRepo)(nil)

// --- helpers ---

func runSync(imp *CSVImporter, jobID, accountID, userID uuid.UUID, data []byte) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		imp.ProcessCSV(jobID, accountID, userID, data)
	}()
	wg.Wait()
}

// --- tests ---

func TestProcessCSV_ValidCSV_JobMarkedDone(t *testing.T) {
	jobID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()

	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n2024-01-15,Whole Foods,-42.50\n2024-01-16,Amazon,-100.00\n")
	runSync(imp, jobID, accountID, userID, data)

	if len(jobs.updateStatusCalls) == 0 || jobs.updateStatusCalls[0].status != "processing" {
		t.Errorf("want first status call to be 'processing', got %+v", jobs.updateStatusCalls)
	}
	if len(jobs.completeCalls) == 0 || jobs.completeCalls[len(jobs.completeCalls)-1].status != "done" {
		t.Errorf("want final complete call to be 'done', got %+v", jobs.completeCalls)
	}
	if len(txs.bulkInsertCalls) == 0 {
		t.Fatal("want at least one BulkInsert call")
	}
	total := 0
	for _, batch := range txs.bulkInsertCalls {
		total += len(batch)
	}
	if total != 2 {
		t.Errorf("want 2 transactions inserted, got %d", total)
	}

	row := txs.bulkInsertCalls[0][0]
	if row.Source != "csv" {
		t.Errorf("want source=csv, got %s", row.Source)
	}
	if row.Classified {
		t.Error("want classified=false")
	}
}

func TestProcessCSV_MMDDYYYYDate_ParsedCorrectly(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n01/15/2024,Coffee,-5.00\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(jobs.completeCalls) == 0 || jobs.completeCalls[0].status != "done" {
		t.Errorf("want done, got %+v", jobs.completeCalls)
	}
	if len(txs.bulkInsertCalls) == 0 || len(txs.bulkInsertCalls[0]) != 1 {
		t.Fatal("want 1 transaction inserted")
	}
	want := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := txs.bulkInsertCalls[0][0].Date
	if !got.Equal(want) {
		t.Errorf("want date %v, got %v", want, got)
	}
}

func TestProcessCSV_DollarSignAndCommasInAmount_ParsedCorrectly(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n2024-01-15,Paycheck,\"$1,234.56\"\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(txs.bulkInsertCalls) == 0 || len(txs.bulkInsertCalls[0]) != 1 {
		t.Fatal("want 1 transaction inserted")
	}
	if txs.bulkInsertCalls[0][0].Amount != 1234.56 {
		t.Errorf("want 1234.56, got %v", txs.bulkInsertCalls[0][0].Amount)
	}
}

func TestProcessCSV_EmptyCSV_JobDoneZeroRows(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(jobs.completeCalls) == 0 || jobs.completeCalls[0].status != "done" {
		t.Errorf("want done, got %+v", jobs.completeCalls)
	}
	if len(txs.bulkInsertCalls) != 0 {
		t.Errorf("want no BulkInsert calls for empty CSV, got %d", len(txs.bulkInsertCalls))
	}
	if len(jobs.updateRowsTotalCalls) == 0 || jobs.updateRowsTotalCalls[0].total != 0 {
		t.Errorf("want rows_total=0, got %+v", jobs.updateRowsTotalCalls)
	}
}

func TestProcessCSV_MissingRequiredColumn_JobFailed(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description\n2024-01-15,Whole Foods\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(jobs.completeCalls) == 0 || jobs.completeCalls[0].status != "failed" {
		t.Errorf("want failed, got %+v", jobs.completeCalls)
	}
}

func TestProcessCSV_MalformedAmount_RowSkippedRestImported(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n2024-01-15,Bad Row,not-a-number\n2024-01-16,Good Row,-50.00\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(jobs.completeCalls) == 0 || jobs.completeCalls[0].status != "done" {
		t.Errorf("want done, got %+v", jobs.completeCalls)
	}
	total := 0
	for _, batch := range txs.bulkInsertCalls {
		total += len(batch)
	}
	if total != 1 {
		t.Errorf("want 1 transaction (bad row skipped), got %d", total)
	}
}

func TestProcessCSV_UpdateStatusProcessingCalledFirst(t *testing.T) {
	jobs := &mockJobRepo{}
	txs := &mockTxRepo{}
	imp := New(jobs, txs, &mockAccRepo{})

	data := []byte("date,description,amount\n2024-01-15,Test,-10.00\n")
	runSync(imp, uuid.New(), uuid.New(), uuid.New(), data)

	if len(jobs.updateStatusCalls) == 0 {
		t.Fatal("want UpdateStatus called")
	}
	if jobs.updateStatusCalls[0].status != "processing" {
		t.Errorf("want first status call 'processing', got %q", jobs.updateStatusCalls[0].status)
	}
	if len(jobs.completeCalls) == 0 {
		t.Fatal("want Complete called")
	}
	last := jobs.completeCalls[len(jobs.completeCalls)-1]
	if last.status != "done" {
		t.Errorf("want last complete call 'done', got %q", last.status)
	}
}
