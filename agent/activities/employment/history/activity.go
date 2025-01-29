// Package history provides employment history verification activities
package history

import (
	"context"
	"fmt"

	"encore.app/agent/types"
)

// Request represents an employment history verification request
type Request struct {
	EmployerName    string
	EmployerContact string
	Profile         types.ResearchProfile
}

// Result represents the result of an employment history verification
type Result struct {
	Verified bool
	Profile  types.ResearchProfile
}

// Activity handles employment history verification operations
type Activity struct {
	config types.Config
}

// NewActivity creates a new employment history verification activity
func NewActivity(config types.Config) *Activity {
	return &Activity{
		config: config,
	}
}

// CheckEmploymentHistory checks employment history with a previous employer
func (a *Activity) CheckEmploymentHistory(ctx context.Context, req *Request) (*Result, error) {
	if req.EmployerName == "" || req.EmployerContact == "" {
		return nil, fmt.Errorf("employer name and contact information are required")
	}

	// Validate research profile
	if req.Profile.ID == "" || req.Profile.FullName == "" {
		return nil, fmt.Errorf("research profile ID and full name are required")
	}

	// Validate at least one current experience
	if len(req.Profile.CurrentExperiences) == 0 {
		return nil, fmt.Errorf("at least one current experience is required")
	}

	// TODO: Implement actual employment verification
	return &Result{
		Verified: false,
		Profile:  req.Profile,
	}, nil
}
