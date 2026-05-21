package predictor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

func TestClassify_HappyPath(t *testing.T) {
	txID := uuid.New()
	catID := uuid.New()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/classify" {
			t.Errorf("want /classify, got %s", r.URL.Path)
		}

		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, ok := body["transactions"]; !ok {
			t.Error("want transactions field in request")
		}
		if _, ok := body["categories"]; !ok {
			t.Error("want categories field in request")
		}

		resp := classifyResponse{
			Predictions: []classifyRespPrediction{
				{TransactionID: txID, CategoryID: &catID, MerchantName: "Whole Foods"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewHTTPPredictorClient(srv.URL)

	txns := []models.Transaction{
		{ID: txID, Description: "WHOLE FOODS MARKET #123", Amount: -42.50},
	}
	cats := []models.Category{
		{ID: catID, Name: "Groceries"},
	}

	predictions, err := client.Classify(context.Background(), txns, cats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(predictions) != 1 {
		t.Fatalf("want 1 prediction, got %d", len(predictions))
	}
	if predictions[0].TransactionID != txID {
		t.Errorf("want txID %v, got %v", txID, predictions[0].TransactionID)
	}
	if predictions[0].CategoryID == nil || *predictions[0].CategoryID != catID {
		t.Errorf("want catID %v, got %v", catID, predictions[0].CategoryID)
	}
	if predictions[0].MerchantName != "Whole Foods" {
		t.Errorf("want merchant Whole Foods, got %v", predictions[0].MerchantName)
	}
}

func TestClassify_PartialResponse_AbsentTransactionsNotInResult(t *testing.T) {
	txID1 := uuid.New()
	txID2 := uuid.New()
	txID3 := uuid.New()
	catID := uuid.New()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only return predictions for tx1 and tx3 — tx2 absent
		resp := classifyResponse{
			Predictions: []classifyRespPrediction{
				{TransactionID: txID1, CategoryID: &catID, MerchantName: "Merchant A"},
				{TransactionID: txID3, CategoryID: &catID, MerchantName: "Merchant C"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewHTTPPredictorClient(srv.URL)

	txns := []models.Transaction{
		{ID: txID1, Description: "desc1", Amount: -10},
		{ID: txID2, Description: "desc2", Amount: -20},
		{ID: txID3, Description: "desc3", Amount: -30},
	}

	predictions, err := client.Classify(context.Background(), txns, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(predictions) != 2 {
		t.Fatalf("want 2 predictions, got %d", len(predictions))
	}

	ids := map[uuid.UUID]bool{}
	for _, p := range predictions {
		ids[p.TransactionID] = true
	}
	if !ids[txID1] || !ids[txID3] {
		t.Errorf("want predictions for tx1 and tx3, got %v", ids)
	}
	if ids[txID2] {
		t.Error("want tx2 absent from predictions")
	}
}

func TestClassify_HTTPError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewHTTPPredictorClient(srv.URL)

	_, err := client.Classify(context.Background(), []models.Transaction{{ID: uuid.New()}}, nil)
	if err == nil {
		t.Fatal("want error from HTTP 500, got nil")
	}
}

func TestClassify_ParentNameIncludedForSubcategory(t *testing.T) {
	groupID := uuid.New()
	catID := uuid.New()
	txID := uuid.New()

	var receivedBody classifyRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resp := classifyResponse{Predictions: []classifyRespPrediction{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewHTTPPredictorClient(srv.URL)

	groupName := "Food"
	cats := []models.Category{
		{ID: groupID, Name: groupName, ParentID: nil},
		{ID: catID, Name: "Groceries", ParentID: &groupID},
	}
	txns := []models.Transaction{{ID: txID, Description: "store", Amount: -10}}

	_, err := client.Classify(context.Background(), txns, cats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedBody.Categories) != 2 {
		t.Fatalf("want 2 categories in request, got %d", len(receivedBody.Categories))
	}

	var groceriesCat *classifyReqCategory
	for i := range receivedBody.Categories {
		if receivedBody.Categories[i].Name == "Groceries" {
			groceriesCat = &receivedBody.Categories[i]
		}
	}
	if groceriesCat == nil {
		t.Fatal("Groceries category not in request")
	}
	if groceriesCat.ParentName == nil || *groceriesCat.ParentName != "Food" {
		t.Errorf("want parent_name=Food, got %v", groceriesCat.ParentName)
	}
}
