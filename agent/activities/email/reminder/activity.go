// Package reminder provides reminder-specific email activities
package reminder

import (
	"context"
	"fmt"
	"time"

	"encore.app/agent/types"
	"encore.app/email"
)

// Activity handles reminder-specific email operations
type Activity struct{}

// NewActivity creates a new reminder email activity instance
func NewActivity(_ interface{}) *Activity {
	return &Activity{}
}

// SendReminder sends a reminder email
func (a *Activity) SendReminder(ctx context.Context, to, from, subject string, data map[string]interface{}) (*types.Result, error) {
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
		TemplateID:   "screening-reminder-notify-v1",
		TemplateData: templateData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send reminder: %w", err)
	}

	return &types.Result{
		Sent:      true,
		MessageID: resp.MessageID,
		SentAt:    time.Now(),
	}, nil
}
