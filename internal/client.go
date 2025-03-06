package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	return c.PostWithContext(context.Background(), endpoint, body)
}

func (c *HTTPClient) Get(endpoint string) (*http.Response, error) {
	return c.GetWithContext(context.Background(), endpoint)
}

func (c *HTTPClient) PostWithContext(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s", c.baseURL, endpoint), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *HTTPClient) GetWithContext(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", c.baseURL, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.client.Do(req)
}
