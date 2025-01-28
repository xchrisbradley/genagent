package activities

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

var (
	// defaultClient is a shared HTTP client used across all activity executions
	defaultClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

// HTTPActivity maintains shared state across activity executions
type HTTPActivity struct {
	client *http.Client
}

// NewHTTPActivity creates a new HTTP activity that uses the shared client
func NewHTTPActivity() *HTTPActivity {
	return &HTTPActivity{
		client: defaultClient,
	}
}

// RequestParams represents the parameters for an HTTP request
type RequestParams struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Response represents the result of an HTTP request
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// Execute performs an HTTP request
func (a *HTTPActivity) Execute(ctx context.Context, params *RequestParams) (*Response, error) {
	// Create request
	var bodyReader io.Reader
	if params.Body != "" {
		bodyReader = bytes.NewBufferString(params.Body)
	}

	req, err := http.NewRequestWithContext(ctx, params.Method, params.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
	}, nil
}
