package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//encore:api public method=GET path=/policy
func (s *Service) ListPolicy(ctx context.Context, params *ListPolicyParams) (*ListPolicyResponse, error) {
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
	countQuery := "SELECT COUNT(*) FROM policy " + whereClause
	selectQuery := "SELECT id, workflow_id, submitted_date, completed_date, definition FROM policy " + whereClause

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

	// Query the policys
	rows, err := db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %v", err)
	}
	defer rows.Close()

	// Parse the results
	policies := []Policy{}
	var rowPolicys []Policy // Store [olicys before status filtering
	for rows.Next() {
		var p Policy
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
			return nil, fmt.Errorf("failed to scan policy row: %v", err)
		}

		if submittedDate.Valid {
			p.SubmittedDate = submittedDate.Time.Format(time.RFC3339)
		}
		if completedDate.Valid {
			p.CompletedDate = completedDate.Time.Format(time.RFC3339)
		}

		// Parse the policy definition
		if err := json.Unmarshal(defJSON, &p.Definition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policy definition: %v", err)
		}

		rowPolicys = append(rowPolicys, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating policy rows: %v", err)
	}

	// Enrich all policies with status and apply status filter if needed
	for _, p := range rowPolicys {
		if err := s.enrichPolicyWithStatus(ctx, &p); err != nil {
			return nil, err
		}

		// Only include if status matches filter (if provided)
		if params.Status == "" || p.Status == params.Status {
			policies = append(policies, p)
		}
	}

	// Update pagination if filtered
	if params.Status != "" {
		totalItems = len(policies)
		totalPages = (totalItems + params.PageSize - 1) / params.PageSize
		if totalPages == 0 {
			totalPages = 1
		}
	}

	return &ListPolicyResponse{
		Policies:    policies,
		TotalItems:  totalItems,
		CurrentPage: params.Page,
		TotalPages:  totalPages,
	}, nil
}
