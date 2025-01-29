// Package consent provides consent-specific email activities
package consent

import (
	"context"
	"fmt"
	"time"

	"encore.app/email"
)

// Activity handles consent-specific email operations
type Activity struct{}

// NewActivity creates a new consent email activity instance
func NewActivity(_ interface{}) *Activity {
	return &Activity{}
}

type SendEmailConsentParams struct {
	To           string            `json:"to"`
	From         string            `json:"from"`
	Subject      string            `json:"subject"`
	TemplateID   string            `json:"template_id"`
	TemplateData map[string]string `json:"template_data"`
}

type SendEmailConsentResponse struct {
	MessageID string    `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
}

// SendConsentEmail sends the consent email
func (a *Activity) SendConsentEmail(ctx context.Context, to, from, subject string, data map[string]interface{}) (*SendEmailConsentResponse, error) {
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
		TemplateID:   "consent",
		TemplateData: templateData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send consent email: %w", err)
	}

	return &SendEmailConsentResponse{
		MessageID: resp.MessageID,
		SentAt:    time.Now(),
	}, nil
}
