package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type InventoryClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewInventoryClient(baseURL string, httpClient *http.Client) *InventoryClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &InventoryClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

type ReserveResult struct {
	ProductID string `json:"product_id"`
	Reserved  int    `json:"reserved"`
	Remaining int    `json:"remaining"`
}

func (c *InventoryClient) Reserve(ctx context.Context, productID string, quantity int) (*ReserveResult, error) {
	body, _ := json.Marshal(map[string]any{
		"product_id": productID,
		"quantity":   quantity,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/inventory/reserve", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reserving inventory for %s: %w", productID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("inventory reserve failed for %s: %s", productID, errResp.Error)
	}

	var result ReserveResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding reserve response: %w", err)
	}
	return &result, nil
}
