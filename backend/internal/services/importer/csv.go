package importer

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

// ImportJobRepository is the subset of repository.ImportJobRepository used by CSVImporter.
type ImportJobRepository interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateRowsTotal(ctx context.Context, id uuid.UUID, rowsTotal int) error
	IncrementRowsImported(ctx context.Context, id uuid.UUID, count int) error
	Complete(ctx context.Context, id uuid.UUID, status string) error
}

// TransactionRepository is the subset of repository.TransactionRepository used by CSVImporter.
type TransactionRepository interface {
	BulkInsert(ctx context.Context, transactions []repository.CreateTransactionParams) (int, error)
}

// CSVImporter parses CSV files and imports transactions into the database.
type CSVImporter struct {
	jobRepo     ImportJobRepository
	txRepo      TransactionRepository
	accountRepo repository.AccountRepository
}

func New(jobRepo ImportJobRepository, txRepo TransactionRepository, accountRepo repository.AccountRepository) *CSVImporter {
	return &CSVImporter{
		jobRepo:     jobRepo,
		txRepo:      txRepo,
		accountRepo: accountRepo,
	}
}

type csvRow struct {
	date        time.Time
	description string
	amount      float64
}

var dateFmts = []string{"2006-01-02", "01/02/2006", "01/02/06"}

func parseDate(s string) (time.Time, error) {
	for _, layout := range dateFmts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %q", s)
}

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

func parseCSV(data []byte) ([]csvRow, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	dateIdx, descIdx, amtIdx := -1, -1, -1
	for i, h := range headers {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "date":
			dateIdx = i
		case "description":
			descIdx = i
		case "amount":
			amtIdx = i
		}
	}
	if dateIdx == -1 || descIdx == -1 || amtIdx == -1 {
		return nil, fmt.Errorf("CSV must have date, description, and amount columns")
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV rows: %w", err)
	}

	var rows []csvRow
	for _, rec := range records {
		if len(rec) <= dateIdx || len(rec) <= descIdx || len(rec) <= amtIdx {
			continue
		}
		d := strings.TrimSpace(rec[dateIdx])
		desc := strings.TrimSpace(rec[descIdx])
		amt := strings.TrimSpace(rec[amtIdx])
		if d == "" || desc == "" || amt == "" {
			continue
		}
		parsedDate, err := parseDate(d)
		if err != nil {
			continue
		}
		parsedAmt, err := parseAmount(amt)
		if err != nil {
			continue
		}
		rows = append(rows, csvRow{date: parsedDate, description: desc, amount: parsedAmt})
	}
	return rows, nil
}

// ProcessCSV is called as a goroutine. It manages the full job lifecycle and
// logs errors internally without propagating them.
func (s *CSVImporter) ProcessCSV(jobID uuid.UUID, accountID uuid.UUID, userID uuid.UUID, fileContent []byte) {
	ctx := context.Background()

	if err := s.jobRepo.UpdateStatus(ctx, jobID, "processing"); err != nil {
		log.Printf("import job %s: update status processing: %v", jobID, err)
		return
	}

	rows, err := parseCSV(fileContent)
	if err != nil {
		log.Printf("import job %s: parse CSV: %v", jobID, err)
		if completeErr := s.jobRepo.Complete(ctx, jobID, "failed"); completeErr != nil {
			log.Printf("import job %s: complete failed: %v", jobID, completeErr)
		}
		return
	}

	if err := s.jobRepo.UpdateRowsTotal(ctx, jobID, len(rows)); err != nil {
		log.Printf("import job %s: update rows total: %v", jobID, err)
	}

	const batchSize = 100
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		params := make([]repository.CreateTransactionParams, len(batch))
		for j, row := range batch {
			params[j] = repository.CreateTransactionParams{
				UserID:      userID,
				AccountID:   accountID,
				Amount:      row.amount,
				Description: row.description,
				Date:        row.date,
				Source:      "csv",
				Classified:  false,
			}
		}

		n, err := s.txRepo.BulkInsert(ctx, params)
		if err != nil {
			log.Printf("import job %s: bulk insert batch starting at %d: %v", jobID, i, err)
			if completeErr := s.jobRepo.Complete(ctx, jobID, "failed"); completeErr != nil {
				log.Printf("import job %s: complete failed: %v", jobID, completeErr)
			}
			return
		}

		if incErr := s.jobRepo.IncrementRowsImported(ctx, jobID, n); incErr != nil {
			log.Printf("import job %s: increment rows imported: %v", jobID, incErr)
		}
	}

	if len(rows) > 0 {
		totalAmount := 0.0
		for _, row := range rows {
			totalAmount += row.amount
		}
		if err := s.accountRepo.UpdateBalanceDirect(ctx, accountID, totalAmount); err != nil {
			log.Printf("import job %s: update account balance: %v", jobID, err)
		}

		account, err := s.accountRepo.GetByID(ctx, accountID, userID)
		if err != nil {
			log.Printf("import job %s: get account for snapshots: %v", jobID, err)
		} else {
			uniqueDates := make(map[time.Time]struct{})
			for _, row := range rows {
				uniqueDates[row.date] = struct{}{}
			}
			for date := range uniqueDates {
				if upsertErr := s.accountRepo.UpsertBalanceSnapshotDirect(ctx, accountID, date, account.Balance); upsertErr != nil {
					log.Printf("import job %s: upsert snapshot for %s: %v", jobID, date.Format("2006-01-02"), upsertErr)
				}
			}
		}
	}

	if err := s.jobRepo.Complete(ctx, jobID, "done"); err != nil {
		log.Printf("import job %s: complete done: %v", jobID, err)
	}
}
