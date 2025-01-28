package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

//encore:api public method=GET path=/policy/:id
func (s *Service) GetPolicy(ctx context.Context, id int) (*Policy, error) {
  var p Policy
  var defJSON []byte
  var submittedDate, completedDate sql.NullTime

  err := db.QueryRow(ctx, `
	  SELECT id, workflow_id, status, submitted_date, completed_date, definition
	  FROM policy
	  WHERE id = $1
	`, id).Scan(
    &p.ID,
    &p.WorkflowID,
    &p.Status,
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

  if err := s.enrichPolicyWithStatus(ctx, &p); err != nil {
		return nil, err
	}

  return &p, nil
}
