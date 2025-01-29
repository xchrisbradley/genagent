package email

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"encore.dev/rlog"
	mail "github.com/xhit/go-simple-mail/v2"
)

// SendParams defines the parameters for sending an email
type SendParams struct {
	To           string            `json:"to"`
	From         string            `json:"from"`
	Subject      string            `json:"subject"`
	TemplateID   string            `json:"templateId"`
	TemplateData map[string]string `json:"templateData"` // Using string values for simplicity
}

// SendResponse defines the response from sending an email
type SendResponse struct {
	MessageID string `json:"messageId"`
}

// Config holds SMTP configuration
type Config struct {
	Host string
	Port int
	Stub bool
}

// Service is the email service implementation
//
//encore:service
type Service struct {
	cfg       Config
	templates map[string]*templatePair
}

type templatePair struct {
	html *template.Template
	text *template.Template
}

func initService() (*Service, error) {
	// Use default config for now
	cfg := Config{
		Host: "localhost",
		Port: 1025, // Default MailHog port
		Stub: false,
	}

	svc := &Service{
		cfg:       cfg,
		templates: make(map[string]*templatePair),
	}

	// Load templates
	if err := svc.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return svc, nil
}

// loadTemplates loads all templates from the templates directory
func (s *Service) loadTemplates() error {
	templatesDir := "email/templates"
	htmlTemplates := make(map[string]*template.Template)
	textTemplates := make(map[string]*template.Template)

	// Walk through templates directory
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			rlog.Error("error walking path", "path", path, "error", err)
			return err
		}
		if info.IsDir() {
			rlog.Info("skipping directory", "path", path)
			return nil
		}

		rlog.Info("processing file", "path", path)

		// Get relative path from templates dir for template ID
		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			rlog.Error("failed to get relative path", "path", path, "error", err)
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Remove file extension to get template ID
		templateID := strings.TrimSuffix(strings.TrimSuffix(relPath, ".html"), ".go")
		templateID = strings.TrimSuffix(strings.TrimSuffix(templateID, ".tmpl"), ".go")
		rlog.Info("template ID", "id", templateID, "path", path)

		// Parse template based on file suffix
		rlog.Info("checking file", "path", path)
		if strings.HasSuffix(path, ".go.html") {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				rlog.Error("failed to parse HTML template", "path", path, "error", err)
				return fmt.Errorf("failed to parse HTML template %s: %w", path, err)
			}
			rlog.Info("loaded HTML template", "id", templateID, "path", path)
			htmlTemplates[templateID] = tmpl
		} else if strings.HasSuffix(path, ".go.tmpl") {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				rlog.Error("failed to parse text template", "path", path, "error", err)
				return fmt.Errorf("failed to parse text template %s: %w", path, err)
			}
			rlog.Info("loaded text template", "id", templateID, "path", path)
			textTemplates[templateID] = tmpl
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk templates directory: %w", err)
	}

	// Match HTML and text templates
	for id, htmlTmpl := range htmlTemplates {
		rlog.Info("matching templates", "id", id)
		textTmpl, ok := textTemplates[id]
		if !ok {
			rlog.Error("missing text template", "id", id)
			return fmt.Errorf("missing text template for %s", id)
		}
		s.templates[id] = &templatePair{
			html: htmlTmpl,
			text: textTmpl,
		}
		rlog.Info("loaded template pair", "id", id)
	}

	if len(s.templates) == 0 {
		return fmt.Errorf("no templates found in %s", templatesDir)
	}

	return nil
}

// getTemplate returns the template pair for the given template ID
func (s *Service) getTemplate(templateID string) (*templatePair, error) {
	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	return tmpl, nil
}

func (s *Service) sendMail(htmlTmpl *template.Template, textTmpl *template.Template, params *SendParams) error {
	if s.cfg.Stub {
		rlog.Info("SMTP stub mode - not sending email",
			"from", params.From,
			"to", params.To,
			"subject", params.Subject,
			"template", params.TemplateID,
		)
		return nil
	}

	// Initialize template data with defaults
	templateData := map[string]interface{}{
		"LogoURL":    "https://www.dietzgen.com/wp-content/uploads/2020/07/Check-PNG-Transparent-Image.png",
		"LogoAlt":    "Company Logo",
		"Domain":     "http://localhost:5173",
		"Signature":  "Thanks",
		"SystemName": "Background Check System",
	}

	// Override defaults with provided data
	for k, v := range params.TemplateData {
		templateData[k] = v
	}

	// Validate required fields based on template ID
	switch {
	case strings.HasPrefix(params.TemplateID, "screening-reminder"):
		required := []string{"Token", "Email", "Path"}
		defaults := map[string]string{
			"NavTitle":      "Action Required",
			"Greeting":      "Hello",
			"Message":       "This is a reminder that you have a pending action to complete.",
			"Action":        "Continue",
			"ActionMessage": "To complete the action, please visit:",
		}
		for k, v := range defaults {
			if _, ok := templateData[k]; !ok {
				templateData[k] = v
			}
		}
		for _, field := range required {
			if _, ok := templateData[field]; !ok {
				return fmt.Errorf("missing required field for reminder template: %s", field)
			}
		}

	case strings.HasPrefix(params.TemplateID, "screening-consent"):
		required := []string{"Token", "Email"}
		defaults := map[string]string{
			"NavTitle":      "Background Check Request",
			"Step1":         "Background Check Request",
			"Step2":         "Enter Personal Information",
			"Step3":         "Background Check Started",
			"Step4":         "Complete",
			"Greeting":      "Hello",
			"Message":       "Your potential employer has requested that we conduct a background check on their behalf.",
			"Action":        "Continue",
			"ActionMessage": "To accept this check, please visit:",
		}
		for k, v := range defaults {
			if _, ok := templateData[k]; !ok {
				templateData[k] = v
			}
		}
		for _, field := range required {
			if _, ok := templateData[field]; !ok {
				return fmt.Errorf("missing required field for consent template: %s", field)
			}
		}

	case strings.HasPrefix(params.TemplateID, "screening-verification"):
		required := []string{"Token", "Email"}
		defaults := map[string]string{
			"NavTitle":      "Employment Verification",
			"Step1":         "Start Employment Verification",
			"Step2":         "Confirm Employment Verification",
			"Step3":         "Verification Submitted",
			"Greeting":      "Hello",
			"Message":       "A candidate is undergoing a background check and we need to verify their current employer.",
			"Action":        "Verify",
			"ActionMessage": "To verify their employment history, please visit:",
		}
		for k, v := range defaults {
			if _, ok := templateData[k]; !ok {
				templateData[k] = v
			}
		}
		for _, field := range required {
			if _, ok := templateData[field]; !ok {
				return fmt.Errorf("missing required field for verification template: %s", field)
			}
		}
	}

	// Create HTML body
	var htmlBody bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBody, templateData); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	// Create text body
	var textBody bytes.Buffer
	if err := textTmpl.Execute(&textBody, templateData); err != nil {
		return fmt.Errorf("failed to execute text template: %w", err)
	}

	// Create SMTP client
	server := mail.NewSMTPClient()
	server.Host = s.cfg.Host
	server.Port = s.cfg.Port
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	// Create email
	email := mail.NewMSG()
	email.SetFrom(params.From).
		AddTo(params.To).
		SetSubject(params.Subject).
		SetBody(mail.TextPlain, textBody.String()).
		AddAlternative(mail.TextHTML, htmlBody.String())

	// Send email
	if err := email.Send(smtpClient); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	rlog.Info("Email sent successfully",
		"from", params.From,
		"to", params.To,
		"subject", params.Subject,
		"template", params.TemplateID,
	)
	return nil
}
