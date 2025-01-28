package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"encore.dev/storage/sqldb"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

type Pipeline struct {
	ID            int                `json:"id"`
	WorkflowID    string             `json:"workflowID"`
	Status        string             `json:"status"`
	SubmittedDate string             `json:"submittedDate"`
	CompletedDate string             `json:"completedDate,omitempty"`
	Definition    PipelineDefinition `json:"definition"`
}

// enrichPipelineWithStatus adds Temporal workflow status to the pipeline
func (s *Service) enrichPipelineWithStatus(ctx context.Context, p *Pipeline) error {
	status, err := s.GetStatus(ctx, p.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow status: %w", err)
	}
	p.Status = status
	return nil
}

// GetStatus fetches and converts the current workflow status from Temporal
func (s *Service) GetStatus(ctx context.Context, workflowID string) (string, error) {
	workflow := s.client.GetWorkflow(ctx, workflowID, "")

	// Get the current run
	run := workflow.GetRunID()
	desc, err := s.client.DescribeWorkflowExecution(ctx, workflowID, run)
	if err != nil {
		// Check specific error types for proper handling
		var applicationErr *temporal.ApplicationError
		if errors.As(err, &applicationErr) {
			// This is an application-level error, workflow should be marked as failed
			return "FAILED", nil
		}

		// For other errors, we should retry the workflow task
		return "RUNNING", nil // Assume it's still running if we can't determine status
	}

	// Convert Temporal's enum values to our string status
	switch desc.WorkflowExecutionInfo.Status {
	case enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED:
		return "RUNNING", nil // Default to running if unspecified
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return "RUNNING", nil
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return "COMPLETED", nil
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return "FAILED", nil
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return "CANCELLED", nil
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return "TERMINATED", nil
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return "CONTINUED_AS_NEW", nil
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return "TIMED_OUT", nil
	default:
		return fmt.Sprintf("UNKNOWN(%d)", desc.WorkflowExecutionInfo.Status), nil
	}
}

type ListPipelineResponse struct {
	Pipelines   []Pipeline `json:"pipelines"`
	TotalItems  int        `json:"totalItems"`
	CurrentPage int        `json:"currentPage"`
	TotalPages  int        `json:"totalPages"`
}

type ListPipelineParams struct {
	Page            int    `query:"page"`
	PageSize        int    `query:"pageSize"`
	Status          string `query:"status"`
	Search          string `query:"search"`
	SubmittedAfter  string `query:"submittedAfter"`
	SubmittedBefore string `query:"submittedBefore"`
	CompletedAfter  string `query:"completedAfter"`
	CompletedBefore string `query:"completedBefore"`
}

// Define database
var db = sqldb.NewDatabase("pipeline", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

//encore:api public method=POST path=/pipeline
func (s *Service) CreatePipeline(ctx context.Context, def PipelineDefinition) (*Pipeline, error) {
	// Generate unique pipeline ID
	pipelineID := fmt.Sprintf("pipeline_%s_%s_%s",
		def.Name,
		def.Version,
		time.Now().Format("20060102150405"))

	// Convert definition to JSON for storage
	defJSON, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pipeline definition: %w", err)
	}

	// Start pipeline execution
	options := client.StartWorkflowOptions{
		ID:        pipelineID,
		TaskQueue: pipelineTaskQueue,
	}

	execution, err := s.client.ExecuteWorkflow(ctx, options, ExecutePipeline, def)
	if err != nil {
		return nil, fmt.Errorf("failed to start pipeline: %w", err)
	}

	// Create pipeline record
	pipeline := &Pipeline{
		WorkflowID:    execution.GetID(),
		SubmittedDate: time.Now().Format(time.RFC3339),
		Definition:    def,
	}

	// Insert into database
	query := `
		INSERT INTO pipeline (workflow_id, submitted_date, definition)
		VALUES ($1, $2, $3)
		RETURNING id`

	err = db.QueryRow(ctx, query,
		pipeline.WorkflowID,
		pipeline.SubmittedDate,
		defJSON,
	).Scan(&pipeline.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline record: %w", err)
	}

	// Get initial status from Temporal
	if err := s.enrichPipelineWithStatus(ctx, pipeline); err != nil {
		return nil, err
	}
	return pipeline, nil
}

//encore:api public method=GET path=/pipeline
func (s *Service) ListPipeline(ctx context.Context, params *ListPipelineParams) (*ListPipelineResponse, error) {
	// Validate page and pageSize
	if params.Page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if params.PageSize < 1 {
		return nil, fmt.Errorf("pageSize must be greater than 0")
	}

	// Build the query conditions and args
	conditions := []string{"1=1"}
	args := []interface{}{}
	if params.Search != "" {
		conditions = append(conditions, "(workflow_id ILIKE $"+fmt.Sprint(len(args)+1)+" OR definition->>'name' ILIKE $"+fmt.Sprint(len(args)+1)+")")
		args = append(args, "%"+params.Search+"%")
	}
	if params.SubmittedAfter != "" {
		conditions = append(conditions, "submitted_date >= $"+fmt.Sprint(len(args)+1))
		args = append(args, params.SubmittedAfter)
	}
	if params.SubmittedBefore != "" {
		conditions = append(conditions, "submitted_date <= $"+fmt.Sprint(len(args)+1))
		args = append(args, params.SubmittedBefore)
	}
	if params.CompletedAfter != "" {
		conditions = append(conditions, "completed_date >= $"+fmt.Sprint(len(args)+1))
		args = append(args, params.CompletedAfter)
	}
	if params.CompletedBefore != "" {
		conditions = append(conditions, "completed_date <= $"+fmt.Sprint(len(args)+1))
		args = append(args, params.CompletedBefore)
	}

	// Combine conditions
	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Build the final queries
	countQuery := "SELECT COUNT(*) FROM pipeline " + whereClause
	selectQuery := "SELECT id, workflow_id, submitted_date, completed_date, definition FROM pipeline " + whereClause

	// Get total count for pagination
	var totalItems int
	err := db.QueryRow(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %v", err)
	}

	// Calculate pagination
	totalPages := (totalItems + params.PageSize - 1) / params.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Validate page number against total pages
	if params.Page > totalPages {
		return nil, fmt.Errorf("page %d exceeds total pages %d", params.Page, totalPages)
	}

	// Add pagination to select query
	offset := (params.Page - 1) * params.PageSize
	selectQuery += " ORDER BY submitted_date DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, params.PageSize, offset)

	// Query the pipelines
	rows, err := db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pipelines: %v", err)
	}
	defer rows.Close()

	// Parse the results
	pipelines := []Pipeline{}
	var rowPipelines []Pipeline // Store pipelines before status filtering
	for rows.Next() {
		var p Pipeline
		var defJSON []byte
		var submittedDate, completedDate sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.WorkflowID,
			&submittedDate,
			&completedDate,
			&defJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pipeline row: %v", err)
		}

		if submittedDate.Valid {
			p.SubmittedDate = submittedDate.Time.Format(time.RFC3339)
		}
		if completedDate.Valid {
			p.CompletedDate = completedDate.Time.Format(time.RFC3339)
		}

		// Parse the pipeline definition
		if err := json.Unmarshal(defJSON, &p.Definition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pipeline definition: %v", err)
		}

		rowPipelines = append(rowPipelines, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pipeline rows: %v", err)
	}

	// Enrich all pipelines with status and apply status filter if needed
	for _, p := range rowPipelines {
		if err := s.enrichPipelineWithStatus(ctx, &p); err != nil {
			return nil, err
		}

		// Only include if status matches filter (if provided)
		if params.Status == "" || p.Status == params.Status {
			pipelines = append(pipelines, p)
		}
	}

	// Update pagination if filtered
	if params.Status != "" {
		totalItems = len(pipelines)
		totalPages = (totalItems + params.PageSize - 1) / params.PageSize
		if totalPages == 0 {
			totalPages = 1
		}
	}

	return &ListPipelineResponse{
		Pipelines:   pipelines,
		TotalItems:  totalItems,
		CurrentPage: params.Page,
		TotalPages:  totalPages,
	}, nil
}
