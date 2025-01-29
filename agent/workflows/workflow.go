package workflows

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	aconsent "encore.app/agent/activities/email/consent"
	"encore.app/agent/types"
	"encore.app/agent/utils"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ScreeningWorkflowInput represents the input parameters for the screening workflow
type ScreeningWorkflowInput struct {
	JobID            string
	Email            string
	Tier             string
	CurrentEmployer  string // Company requesting the check
	PreviousEmployer string // Company to verify employment with
}

// ScreeningStatus represents the current status of a screening check
type ScreeningStatus struct {
	Status          types.Status
	CompletedSteps  []string
	RemainingSteps  []string
	ConsentReceived bool
	CandidateInfo   *types.CandidateDetails
}

// DefaultActivityOptions returns the default activity options with retry policy
func DefaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
}

// ValidateEmail checks if the provided email is valid
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("recipient email cannot be empty")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid recipient email format: %s", email)
	}

	return nil
}

// GenerateToken is a wrapper around utils.GenerateWorkflowToken that includes logging
func GenerateToken(ctx workflow.Context, purpose string) (string, error) {
	logger := workflow.GetLogger(ctx)
	token, err := utils.GenerateWorkflowToken(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate %s token", purpose), "error", err)
		return "", fmt.Errorf("failed to generate %s token: %w", purpose, err)
	}
	return token, nil
}

// WaitForResearch waits for research signal with a timeout
func WaitForResearch(ctx workflow.Context) (*types.ResearchSubmissionSignal, error) {
	var researchSignal *types.ResearchSubmissionSignal
	var signalErr error

	s := workflow.NewSelector(ctx)
	ch := workflow.GetSignalChannel(ctx, types.ResearchSubmissionSignalName)

	s.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		var signal types.ResearchSubmissionSignal
		c.Receive(ctx, &signal)
		researchSignal = &signal
		workflow.GetLogger(ctx).Info("Received research response",
			"verified", signal.Verified,
			"fullName", signal.Profile.FullName,
			"currentExperience", len(signal.Profile.CurrentExperiences) > 0,
		)
	})

	s.AddFuture(workflow.NewTimer(ctx, types.ResearchGracePeriod), func(f workflow.Future) {
		signalErr = f.Get(ctx, nil)
		researchSignal = nil
		workflow.GetLogger(ctx).Info("Research grace period expired")
	})

	s.Select(ctx)
	if signalErr != nil {
		workflow.GetLogger(ctx).Error("Error waiting for research", "error", signalErr)
		return nil, fmt.Errorf("error waiting for research: %w", signalErr)
	}

	return researchSignal, nil
}

// WaitForAcceptance waits for acceptance signal with a timeout
func WaitForAcceptance(ctx workflow.Context) (bool, error) {
	accepted := false
	var signalErr error

	s := workflow.NewSelector(ctx)
	ch := workflow.GetSignalChannel(ctx, types.AcceptSubmissionSignalName)

	s.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		var signal types.AcceptSubmissionSignal
		c.Receive(ctx, &signal)
		accepted = signal.Accepted
		workflow.GetLogger(ctx).Info("Received consent response",
			"accepted", accepted,
			"fullName", signal.CandidateDetails.FullName,
			"employer", signal.CandidateDetails.CurrentEmployer,
		)
	})

	s.AddFuture(workflow.NewTimer(ctx, types.AcceptGracePeriod), func(f workflow.Future) {
		signalErr = f.Get(ctx, nil)
		accepted = false
		workflow.GetLogger(ctx).Info("Consent grace period expired")
	})

	s.Select(ctx)
	if signalErr != nil {
		workflow.GetLogger(ctx).Error("Error waiting for consent", "error", signalErr)
		return false, fmt.Errorf("error waiting for consent: %w", signalErr)
	}

	return accepted, nil
}

// Screening is the main workflow function
func Screening(ctx workflow.Context, input *ScreeningWorkflowInput) (*ScreeningStatus, error) {
	logger := workflow.GetLogger(ctx)
	status := &ScreeningStatus{
		Status:          types.StatusPending,
		CompletedSteps:  []string{},
		RemainingSteps:  []string{"consent", "background_check", "research"},
		ConsentReceived: false,
	}

	// Register query handler for getting workflow status
	err := workflow.SetQueryHandler(ctx, types.ScreeningStatusQuery, func() (*ScreeningStatus, error) {
		return status, nil
	})
	if err != nil {
		logger.Error("Failed to register query handler", "error", err)
		status.Status = types.StatusFailed
		return status, fmt.Errorf("failed to register query handler: %w", err)
	}

	// Validate email before starting workflow
	if err := ValidateEmail(input.Email); err != nil {
		logger.Error("Invalid email address", "error", err)
		status.Status = types.StatusFailed
		return status, err
	}

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, DefaultActivityOptions())

	// Generate unique token for consent email
	consentToken, err := GenerateToken(ctx, "consent")
	if err != nil {
		status.Status = types.StatusFailed
		return status, err
	}

	logger.Info("Sending screening consent email", "recipient", input.Email)
	req := &aconsent.SendEmailConsentParams{
		To:      input.Email,
		From:    types.CandidateSupportEmail,
		Subject: "Screening Consent Required",
		TemplateData: map[string]string{
			"Token":    consentToken,
			"Email":    input.Email,
			"Employer": input.CurrentEmployer,
			"Domain":   "http://localhost:5173",
			"Name":     input.Email, // Default to email until we get candidate details
		},
	}

	consentActivity := aconsent.NewActivity(nil)
	if err := workflow.ExecuteActivity(ctx, consentActivity.SendConsentEmail, req.To, req.From, req.Subject, req.TemplateData).Get(ctx, nil); err != nil {
		logger.Error("Failed to send consent email", "error", err)
		status.Status = types.StatusFailed
		return status, fmt.Errorf("failed to send consent email: %w", err)
	}

	// Wait for consent response
	var candidateInfo *types.CandidateDetails
	s := workflow.NewSelector(ctx)
	ch := workflow.GetSignalChannel(ctx, types.AcceptSubmissionSignalName)
	accepted := false
	var signalErr error

	s.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		var signal types.AcceptSubmissionSignal
		c.Receive(ctx, &signal)
		accepted = signal.Accepted
		candidateInfo = &signal.CandidateDetails
		workflow.GetLogger(ctx).Info("Received consent response",
			"accepted", accepted,
			"fullName", signal.CandidateDetails.FullName,
			"employer", signal.CandidateDetails.CurrentEmployer,
		)
	})

	s.AddFuture(workflow.NewTimer(ctx, types.AcceptGracePeriod), func(f workflow.Future) {
		signalErr = f.Get(ctx, nil)
		accepted = false
		workflow.GetLogger(ctx).Info("Consent grace period expired")
	})

	s.Select(ctx)
	if signalErr != nil {
		workflow.GetLogger(ctx).Error("Error waiting for consent", "error", signalErr)
		status.Status = types.StatusFailed
		return status, fmt.Errorf("error waiting for consent: %w", signalErr)
	}

	status.ConsentReceived = accepted
	status.CandidateInfo = candidateInfo
	if !accepted {
		logger.Info("Consent not received or declined")
		status.Status = types.StatusDeclined
		return status, nil
	}

	// TODO: Implement remaining screening workflow steps

	logger.Info("Screening workflow completed successfully")
	return status, nil
}
