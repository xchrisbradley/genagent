package email

import (
	"context"
	"fmt"

	"encore.dev/beta/errs"
)

// Send sends an email
//
//encore:api public
func (s *Service) Send(ctx context.Context, params *SendParams) (*SendResponse, error) {
	if params.To == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "recipient email is required"}
	}
	if params.From == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "sender email is required"}
	}
	if params.Subject == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "subject is required"}
	}
	if params.TemplateID == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "template ID is required"}
	}

	// Get template pair for the given template ID
	tmpl, err := s.getTemplate(params.TemplateID)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: fmt.Sprintf("invalid template ID: %v", err)}
	}

	if err := s.sendMail(tmpl.html, tmpl.text, params); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: fmt.Sprintf("failed to send email: %v", err)}
	}

	return &SendResponse{
		MessageID: fmt.Sprintf("%s-%s", params.TemplateID, params.To),
	}, nil
}
