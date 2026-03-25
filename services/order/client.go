package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string, httpClient *http.Client) *ProductClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ProductClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/products/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching product %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned %d for product %s", resp.StatusCode, id)
	}

	var product model.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("decoding product: %w", err)
	}
	return &product, nil
}
