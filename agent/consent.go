package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"encore.app/agent/types"
	"encore.app/agent/utils"
	"encore.dev/rlog"
)

// ConsentResponse represents the API response for consent endpoints
type ConsentResponse struct {
	Email           string `json:"email"`
	CurrentEmployer string `json:"currentEmployer"`
}

// ConsentSubmission represents the user's consent submission
type ConsentSubmission struct {
	Accepted         bool   `json:"accepted"`
	FullName         string `json:"fullName"`
	SSN              string `json:"ssn"`
	Address          string `json:"address"`
	PhoneNumber      string `json:"phoneNumber"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	CurrentEmployer  string `json:"currentEmployer"`
	PreviousEmployer string `json:"previousEmployer"`
}

// ServeConsentAPI handles the consent API endpoints
//
//encore:api public raw path=/api/consent/*token
func (s *Service) ServeConsentAPI(w http.ResponseWriter, req *http.Request) {
	// Enable CORS for development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract and validate token
	path := strings.TrimPrefix(req.URL.Path, "/api/consent/")
	token := strings.TrimSuffix(path, "/")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Parse workflow token
	workflowID, runID, err := utils.ParseWorkflowToken(token)
	if err != nil {
		rlog.Error("Failed to parse workflow token", "error", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Handle GET request for initial consent info
	if req.Method == http.MethodGet {
		email, employer, err := getWorkflowInfo(workflowID, runID, s)
		if err != nil {
			rlog.Error("Failed to get workflow details", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := ConsentResponse{
			Email:           email,
			CurrentEmployer: employer,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			rlog.Error("Failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Handle POST request for consent submission
	if req.Method == http.MethodPost {
		var submission ConsentSubmission
		if err := json.NewDecoder(req.Body).Decode(&submission); err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Validate required fields if consent is accepted
		if submission.Accepted {
			if submission.FullName == "" || submission.SSN == "" || submission.Address == "" ||
				submission.PhoneNumber == "" || submission.PhoneCountryCode == "" {
				http.Error(w, "Missing required information", http.StatusBadRequest)
				return
			}
		}

		// Signal the workflow with the candidate's response
		signal := &types.AcceptSubmissionSignal{
			Accepted: submission.Accepted,
			CandidateDetails: types.CandidateDetails{
				FullName:         submission.FullName,
				SSN:              submission.SSN,
				Address:          submission.Address,
				PhoneNumber:      submission.PhoneNumber,
				PhoneCountryCode: submission.PhoneCountryCode,
				PreviousEmployer: submission.PreviousEmployer,
				CurrentEmployer:  submission.CurrentEmployer,
			},
		}

		if err := s.Client().SignalWorkflow(context.Background(), workflowID, runID, types.AcceptSubmissionSignalName, signal); err != nil {
			http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func getWorkflowInfo(workflowID, runID string, s *Service) (string, string, error) {
	resp, err := s.Client().DescribeWorkflowExecution(context.Background(), workflowID, runID)
	if err != nil {
		return "", "", err
	}

	var email, employer string
	if memo := resp.WorkflowExecutionInfo.GetMemo(); memo != nil && len(memo.GetFields()) > 0 {
		if emailField, ok := memo.GetFields()["Email"]; ok {
			email = string(emailField.GetData())
		}
		if employerField, ok := memo.GetFields()["CurrentEmployer"]; ok {
			employer = string(employerField.GetData())
		}
	}

	// Fallback to workflow ID if email is not available
	if email == "" {
		email = workflowID
	}

	// Strip quotes if present
	email = strings.Trim(email, "\"")
	employer = strings.Trim(employer, "\"")

	return email, employer, nil
}
