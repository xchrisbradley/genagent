package nodes

import (
	"encoding/json"
	"fmt"
	"time"

	"encore.app/pipeline/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// HTTPRequest represents a single HTTP request configuration
type HTTPRequest struct {
	URL     string            `json:"url,omitempty"`    // Optional override
	Method  string            `json:"method,omitempty"` // Optional override
	Headers map[string]string `json:"Headers"`          // Note: Capital H to match input
	Body    string            `json:"Body"`             // Note: Capital B to match input
}

// HTTPNodeConfig represents the configuration for an HTTP node
type HTTPNodeConfig struct {
	URL      string        `json:"url"`      // Default URL for all requests
	Method   string        `json:"method"`   // Default method for all requests
	Requests []HTTPRequest `json:"requests"` // Array of request configurations
}

// Validate ensures the HTTP node configuration is valid
func (c *HTTPNodeConfig) Validate() error {
	if len(c.Requests) == 0 {
		// If no requests array, we need a default URL
		if c.URL == "" {
			return fmt.Errorf("either url or requests array is required")
		}
	} else {
		// For each request, ensure it either has a URL or the node has a default URL
		for i, req := range c.Requests {
			if req.URL == "" && c.URL == "" {
				return fmt.Errorf("request %d missing url and no default url configured", i)
			}
		}
	}
	return nil
}

// HTTPNodeExecutor implements the Executor interface for HTTP nodes
type HTTPNodeExecutor struct {
	httpActivity *activities.HTTPActivity
}

// NewHTTPNodeExecutor creates a new HTTP node executor
func NewHTTPNodeExecutor() *HTTPNodeExecutor {
	return &HTTPNodeExecutor{
		httpActivity: activities.NewHTTPActivity(),
	}
}

// Execute runs the HTTP node with the given configuration
func (e *HTTPNodeExecutor) Execute(ctx workflow.Context, config json.RawMessage) (*NodeResult, error) {
	// Parse config
	var nodeConfig HTTPNodeConfig
	if err := json.Unmarshal(config, &nodeConfig); err != nil {
		// Try parsing as array of requests for backward compatibility
		var requests []HTTPRequest
		if err := json.Unmarshal(config, &requests); err != nil {
			return nil, fmt.Errorf("invalid http node config: %w", err)
		}
		nodeConfig.Requests = requests
	}

	// Validate config
	if err := nodeConfig.Validate(); err != nil {
		return nil, err
	}

	// If no requests array but we have a default URL, create a single request
	if len(nodeConfig.Requests) == 0 && nodeConfig.URL != "" {
		nodeConfig.Requests = []HTTPRequest{{
			URL:     nodeConfig.URL,
			Method:  nodeConfig.Method,
			Headers: make(map[string]string),
		}}
	}

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Execute each request
	var results []map[string]any
	startTime := time.Now()

	for _, req := range nodeConfig.Requests {
		// Merge request config with node defaults
		url := req.URL
		if url == "" {
			url = nodeConfig.URL
		}
		method := req.Method
		if method == "" {
			method = nodeConfig.Method
			if method == "" {
				method = "GET" // Default to GET if not specified
			}
		}

		var httpResp activities.Response
		err := workflow.ExecuteActivity(ctx, e.httpActivity.Execute, &activities.RequestParams{
			URL:     url,
			Method:  method,
			Headers: req.Headers,
			Body:    req.Body,
		}).Get(ctx, &httpResp)

		if err != nil {
			return &NodeResult{
				Success: false,
				Error:   fmt.Sprintf("http activity failed: %v", err),
			}, nil
		}

		results = append(results, map[string]any{
			"statusCode": httpResp.StatusCode,
			"headers":    httpResp.Headers,
			"body":       httpResp.Body,
		})
	}

	// Return combined results
	data := map[string]any{
		"results":       results,
		"executionTime": time.Since(startTime),
	}

	// Determine overall success - all requests must succeed
	success := true
	var errMsg string
	for i, result := range results {
		statusCode := result["statusCode"].(int)
		if statusCode < 200 || statusCode >= 300 {
			success = false
			errMsg = fmt.Sprintf("request %d failed with status code %d", i, statusCode)
			break
		}
	}

	return &NodeResult{
		Success: success,
		Data:    data,
		Error:   errMsg,
	}, nil
}

// ValidateConfig validates the HTTP node configuration
func (e *HTTPNodeExecutor) ValidateConfig(config json.RawMessage) error {
	var nodeConfig HTTPNodeConfig
	if err := json.Unmarshal(config, &nodeConfig); err != nil {
		// Try parsing as array of requests
		var requests []HTTPRequest
		if err := json.Unmarshal(config, &requests); err != nil {
			return fmt.Errorf("invalid http node config: %w", err)
		}
		nodeConfig.Requests = requests
	}
	return nodeConfig.Validate()
}
