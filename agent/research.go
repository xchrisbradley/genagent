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

// ResearchResponse represents the API response for research endpoints
type ResearchResponse struct {
	CandidateEmail string `json:"candidateEmail"`
	CandidateName  string `json:"candidateName"`
	EmployerName   string `json:"employerName"`
}

// ResearchSubmission represents the research data submission
type ResearchSubmission struct {
	Verified             bool                   `json:"verified"`
	ID                   string                 `json:"id"`
	CanonicalURL         string                 `json:"canonical"`
	FullName             string                 `json:"fullName"`
	Headline             string                 `json:"headline,omitempty"`
	Location             types.Location         `json:"location"`
	Industry             string                 `json:"industry,omitempty"`
	ProfilePictureURL    string                 `json:"pictureUrl,omitempty"`
	Social               types.SocialMediaLinks `json:"social"`
	CurrentExperiences   []types.Experience     `json:"currentExperiences"`
	PreviousExperiences  []types.Experience     `json:"previousExperiences"`
	Education            []types.Education      `json:"education"`
	Skills               []string               `json:"skills"`
	About                string                 `json:"about,omitempty"`
	ExperienceInDays     int                    `json:"experienceInDays"`
	HasMilitaryService   bool                   `json:"hasMilitaryService"`
	HasSecurityClearance bool                   `json:"hasSecurityClearance"`
}

// ServeResearchAPI handles the research API endpoints
//
//encore:api public raw path=/api/research/*token
func (s *Service) ServeResearchAPI(w http.ResponseWriter, req *http.Request) {
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
	path := strings.TrimPrefix(req.URL.Path, "/api/research/")
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

	// Handle GET request for research info
	if req.Method == http.MethodGet {
		candidateEmail, candidateName, employerName, err := getResearchInfo(workflowID, runID, s)
		if err != nil {
			rlog.Error("Failed to get research details", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := ResearchResponse{
			CandidateEmail: candidateEmail,
			CandidateName:  candidateName,
			EmployerName:   employerName,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			rlog.Error("Failed to encode response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Handle POST request for research submission
	if req.Method == http.MethodPost {
		var submission ResearchSubmission
		if err := json.NewDecoder(req.Body).Decode(&submission); err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Validate required fields if verified
		if submission.Verified {
			if submission.ID == "" || submission.FullName == "" {
				http.Error(w, "Missing required profile information", http.StatusBadRequest)
				return
			}

			// Validate at least one current or previous experience
			if len(submission.CurrentExperiences) == 0 && len(submission.PreviousExperiences) == 0 {
				http.Error(w, "At least one work experience is required", http.StatusBadRequest)
				return
			}
		}

		// Signal the workflow with the research profile
		signal := &types.ResearchSubmissionSignal{
			Verified: submission.Verified,
			Profile: types.ResearchProfile{
				ID:                   submission.ID,
				CanonicalURL:         submission.CanonicalURL,
				FullName:             submission.FullName,
				Headline:             submission.Headline,
				Location:             submission.Location,
				Industry:             submission.Industry,
				ProfilePictureURL:    submission.ProfilePictureURL,
				Social:               submission.Social,
				CurrentExperiences:   submission.CurrentExperiences,
				PreviousExperiences:  submission.PreviousExperiences,
				Education:            submission.Education,
				Skills:               submission.Skills,
				About:                submission.About,
				ExperienceInDays:     submission.ExperienceInDays,
				HasMilitaryService:   submission.HasMilitaryService,
				HasSecurityClearance: submission.HasSecurityClearance,
			},
		}

		if err := s.Client().SignalWorkflow(context.Background(), workflowID, runID, types.ResearchSubmissionSignalName, signal); err != nil {
			http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func getResearchInfo(workflowID, runID string, s *Service) (string, string, string, error) {
	resp, err := s.Client().DescribeWorkflowExecution(context.Background(), workflowID, runID)
	if err != nil {
		return "", "", "", err
	}

	var candidateEmail, candidateName, employerName string
	if memo := resp.WorkflowExecutionInfo.GetMemo(); memo != nil && len(memo.GetFields()) > 0 {
		if emailField, ok := memo.GetFields()["CandidateEmail"]; ok {
			candidateEmail = string(emailField.GetData())
		}
		if nameField, ok := memo.GetFields()["CandidateName"]; ok {
			candidateName = string(nameField.GetData())
		}
		if employerField, ok := memo.GetFields()["EmployerName"]; ok {
			employerName = string(employerField.GetData())
		}
	}

	// Fallback to workflow ID if email is not available
	if candidateEmail == "" {
		candidateEmail = workflowID
	}

	// Strip quotes if present
	candidateEmail = strings.Trim(candidateEmail, "\"")
	candidateName = strings.Trim(candidateName, "\"")
	employerName = strings.Trim(employerName, "\"")

	return candidateEmail, candidateName, employerName, nil
}
