package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

//encore:api auth method=POST path=/policy
func (s *Service) CreatePolicy(ctx context.Context, def PolicyDefinition) (*Policy, error) {
	// Generate unique Policy ID
	policyID := fmt.Sprintf("policy_%s_%s_%s",
		def.Name,
		def.Version,
		time.Now().Format("20060102150405"))

	// Convert definition to JSON for storage
	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy definition: %w", err)
	}

	// // Start policy execution
	// options := client.StartWorkflowOptions{
	// 	ID:        policyID,
	// 	TaskQueue: policyTaskQueue,
	// }

	// execution, err := s.client.ExecuteWorkflow(ctx, options, ExecutePolicy, def)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to start Policy: %w", err)
	// }

	// Create Policy record
	policy := &Policy{
		WorkflowID:    policyID,
		SubmittedDate: time.Now().Format(time.RFC3339),
		Definition:    def,
	}

	// Get initial status from Temporal
	if err := s.enrichPolicyWithStatus(ctx, policy); err != nil {
		return nil, err
	}

	// Insert into database
	query := `
		INSERT INTO policy (workflow_id, submitted_date, definition, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err = db.QueryRow(ctx, query,
		policy.WorkflowID,
		policy.SubmittedDate,
		defJSON,
		policy.Status,
	).Scan(&policy.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create policy record: %w", err)
	}

	return policy, nil
}
