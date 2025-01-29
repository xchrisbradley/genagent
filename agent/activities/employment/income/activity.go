// Package income provides income verification activities
package income

import (
	"context"
	"fmt"
	"time"

	"encore.app/agent"
)

// Request represents an income verification request
type Request struct {
	EmployerName    string
	EmployerContact string
	Period          string // annual, monthly, etc.
	Year            int
}

// Result represents the result of an income verification
type Result struct {
	Verified   bool
	VerifiedAt time.Time
	VerifiedBy string
	Notes      string
	Income     float64
	Currency   string
	Period     string
	Year       int
}

// Activity handles income verification operations
type Activity struct {
	config agent.Config
}

// NewActivity creates a new income verification activity
func NewActivity(config agent.Config) *Activity {
	return &Activity{
		config: config,
	}
}

// GetIncomeInformation retrieves income information from a previous employer
func (a *Activity) GetIncomeInformation(ctx context.Context, req *Request) (*Result, error) {
	if req.EmployerName == "" || req.EmployerContact == "" {
		return nil, fmt.Errorf("employer name and contact information are required")
	}

	// TODO: Implement actual income verification
	return &Result{
		Verified:   false,
		VerifiedAt: time.Now(),
		VerifiedBy: "system",
		Notes:      "Verification pending",
		Income:     0,
		Currency:   "USD",
		Period:     req.Period,
		Year:       req.Year,
	}, nil
}
