package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPClient is a simple HTTP client.
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewClient creates a new HTTPClient with the given baseURL.
func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

// Get sends a GET request to the given endpoint.
func (c *HTTPClient) Get(endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)
	return c.Client.Get(url)
}

// Post sends a POST request to the given endpoint with the given body.
func (c *HTTPClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	response, err := c.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	return response, err
}
