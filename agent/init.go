package agent

import (
	"context"
	"fmt"

	"encore.app/agent/types"
	"encore.app/agent/workflows"
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
)

// Init starts a new screening workflow for a candidate
//
//encore:api public
func (s *Service) Init(ctx context.Context, req *workflows.ScreeningWorkflowInput) (*workflows.ScreeningStatus, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}
	if req.JobID == "" {
		return nil, fmt.Errorf("jobID is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.Tier == "" {
		return nil, fmt.Errorf("tier is required")
	}
	if req.CurrentEmployer == "" {
		return nil, fmt.Errorf("current employer is required")
	}
	if req.PreviousEmployer == "" {
		return nil, fmt.Errorf("previous employer is required")
	}

	// Generate a unique workflow ID
	workflowID := fmt.Sprintf("agent.init.v%s", req.JobID)

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: GetTaskQueue(),
		Memo: map[string]interface{}{
			"Email":           req.Email,
			"CurrentEmployer": req.CurrentEmployer,
		},
	}

	input := &workflows.ScreeningWorkflowInput{
		JobID:            req.JobID,
		Email:            req.Email,
		Tier:             req.Tier,
		CurrentEmployer:  req.CurrentEmployer,
		PreviousEmployer: req.PreviousEmployer,
	}

	we, err := s.Client().ExecuteWorkflow(ctx, options, workflows.Agent, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start screening workflow: %w", err)
	}

	rlog.Info("started screening workflow",
		"workflow_id", we.GetID(),
		"run_id", we.GetRunID(),
		"email", req.Email,
		"tier", req.Tier,
	)

	return &workflows.ScreeningStatus{
		Status:          types.StatusPending,
		CompletedSteps:  []string{},
		RemainingSteps:  []string{},
		ConsentReceived: false,
	}, nil
}
