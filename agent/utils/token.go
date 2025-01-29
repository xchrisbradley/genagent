package utils

import (
	"encoding/base64"
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
)

// GenerateWorkflowToken generates a token based on workflow execution ID and run ID
func GenerateWorkflowToken(ctx workflow.Context) (string, error) {
	info := workflow.GetInfo(ctx)
	token := fmt.Sprintf("%s:%s", info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
	return base64.URLEncoding.EncodeToString([]byte(token)), nil
}

// ParseWorkflowToken extracts workflow execution ID and run ID from a token
func ParseWorkflowToken(token string) (workflowID, runID string, err error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid token format: %w", err)
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token format: expected workflowID:runID")
	}

	return parts[0], parts[1], nil
}
