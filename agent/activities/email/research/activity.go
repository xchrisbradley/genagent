// Package research provides research-specific email activities
package research

import (
	"context"
	"fmt"
	"time"

	"encore.app/agent/types"
	"encore.app/email"
)

// Activity handles research-specific email operations
type Activity struct{}

// NewActivity creates a new research email activity instance
func NewActivity(_ interface{}) *Activity {
	return &Activity{}
}

// SendResearchRequest sends an employment history research request
func (a *Activity) SendResearchRequest(ctx context.Context, to, from, subject string, data map[string]interface{}) (*types.Result, error) {
	// Convert data to string map as required by email service
	templateData := make(map[string]string)
	for k, v := range data {
		templateData[k] = fmt.Sprintf("%v", v)
	}

	// Only pass required data, defaults are handled by email service
	resp, err := email.Send(ctx, &email.SendParams{
		To:           to,
		From:         from,
		Subject:      subject,
		TemplateID:   "screening-research-request-v1",
		TemplateData: templateData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send research request: %w", err)
	}

	return &types.Result{
		Sent:      true,
		MessageID: resp.MessageID,
		SentAt:    time.Now(),
	}, nil
}
