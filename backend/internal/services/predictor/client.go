package predictor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

type Prediction struct {
	TransactionID uuid.UUID
	CategoryID    *uuid.UUID
	MerchantName  string
}

type PredictorClient interface {
	Classify(ctx context.Context, transactions []models.Transaction, categories []models.Category) ([]Prediction, error)
}

type httpPredictorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPPredictorClient(predictionsServiceURL string) PredictorClient {
	return &httpPredictorClient{
		baseURL:    predictionsServiceURL,
		httpClient: &http.Client{},
	}
}

type classifyReqTransaction struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
}

type classifyReqCategory struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ParentName *string   `json:"parent_name"`
}

type classifyRequest struct {
	Transactions []classifyReqTransaction `json:"transactions"`
	Categories   []classifyReqCategory    `json:"categories"`
}

type classifyRespPrediction struct {
	TransactionID uuid.UUID  `json:"transaction_id"`
	CategoryID    *uuid.UUID `json:"category_id"`
	MerchantName  string     `json:"merchant_name"`
}

type classifyResponse struct {
	Predictions []classifyRespPrediction `json:"predictions"`
}

func (c *httpPredictorClient) Classify(ctx context.Context, transactions []models.Transaction, categories []models.Category) ([]Prediction, error) {
	categoryNameByID := make(map[uuid.UUID]string, len(categories))
	for _, cat := range categories {
		categoryNameByID[cat.ID] = cat.Name
	}

	reqTxns := make([]classifyReqTransaction, len(transactions))
	for i, t := range transactions {
		reqTxns[i] = classifyReqTransaction{
			ID:          t.ID,
			Description: t.Description,
			Amount:      t.Amount,
		}
	}

	reqCats := make([]classifyReqCategory, len(categories))
	for i, cat := range categories {
		var parentName *string
		if cat.ParentID != nil {
			if name, ok := categoryNameByID[*cat.ParentID]; ok {
				parentName = &name
			}
		}
		reqCats[i] = classifyReqCategory{
			ID:         cat.ID,
			Name:       cat.Name,
			ParentName: parentName,
		}
	}

	body, err := json.Marshal(classifyRequest{
		Transactions: reqTxns,
		Categories:   reqCats,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal classify request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/classify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create classify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call prediction service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prediction service returned status %d", resp.StatusCode)
	}

	var result classifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode classify response: %w", err)
	}

	predictions := make([]Prediction, 0, len(result.Predictions))
	for _, p := range result.Predictions {
		predictions = append(predictions, Prediction{
			TransactionID: p.TransactionID,
			CategoryID:    p.CategoryID,
			MerchantName:  p.MerchantName,
		})
	}
	return predictions, nil
}
