package policy

import (
	"context"

	"encore.dev"
	"encore.dev/storage/sqldb"
)

// Use an environment-specific task queue so we can use the same
// Temporal Cluster for all cloud environments.
var (
	envName         = encore.Meta().Environment.Name
	policyTaskQueue = envName + "-policy"
)

type Policy struct {
	ID            int              `json:"id"`
	WorkflowID    string           `json:"workflowID"`
	Status        string           `json:"status"`
	SubmittedDate string           `json:"submittedDate"`
	CompletedDate string           `json:"completedDate,omitempty"`
	Definition    PolicyDefinition `json:"definition"`
}

type ListPolicyResponse struct {
	Policies    []Policy `json:"policies"`
	TotalItems  int      `json:"totalItems"`
	CurrentPage int      `json:"currentPage"`
	TotalPages  int      `json:"totalPages"`
}

type ListPolicyParams struct {
	Page            int    `query:"page"`
	PageSize        int    `query:"pageSize"`
	Status          string `query:"status"`
	Search          string `query:"search"`
	SubmittedAfter  string `query:"submittedAfter"`
	SubmittedBefore string `query:"submittedBefore"`
	CompletedAfter  string `query:"completedAfter"`
	CompletedBefore string `query:"completedBefore"`
}

//encore:service
type Service struct {
}

// Define database
var db = sqldb.NewDatabase("policy", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// enrichPolicyWithStatus adds Temporal workflow status to the Policy
func (s *Service) enrichPolicyWithStatus(ctx context.Context, p *Policy) error {
	p.Status = "RUNNING"
	return nil
}

func (s *Service) Shutdown(force context.Context) {
	// Implementation for graceful shutdown
}
