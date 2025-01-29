package types

import "time"

// Common constants
const (
	// AcceptSubmissionSignalName is the name of the signal channel for consent acceptance
	AcceptSubmissionSignalName = "accept-submission"

	// ResearchSubmissionSignalName is the name of the signal channel for research acceptance
	ResearchSubmissionSignalName = "research-submission"

	// VerificationSubmissionSignalName is the name of the signal channel for verification acceptance
	VerificationSubmissionSignalName = "verification-submission"

	// AcceptGracePeriod is the time allowed for accepting the screening check
	AcceptGracePeriod = 72 * time.Hour

	// ResearchGracePeriod is the time allowed for completing research
	ResearchGracePeriod = 72 * time.Hour

	// ScreeningStatusQuery is the name of the query for getting workflow status
	ScreeningStatusQuery = "screening-status"
)

// Status represents the current state of a screening check
type Status string

const (
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusDeclined  Status = "declined"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// CandidateDetails contains information provided by the candidate
type CandidateDetails struct {
	FullName         string
	SSN              string
	Address          string
	PhoneNumber      string
	PhoneCountryCode string
	CurrentEmployer  string // Company requesting the check
	PreviousEmployer string // Company to verify employment with
}

// AcceptSubmissionSignal represents the signal sent when a candidate accepts/declines
type AcceptSubmissionSignal struct {
	Accepted         bool
	CandidateDetails CandidateDetails
}

// SocialMediaLinks contains various social media profile URLs
type SocialMediaLinks struct {
	LinkedIn      string `json:"linkedIn,omitempty"`
	GitHub        string `json:"github,omitempty"`
	Twitter       string `json:"twitter,omitempty"`
	Facebook      string `json:"facebook,omitempty"`
	Instagram     string `json:"instagram,omitempty"`
	YouTube       string `json:"youtube,omitempty"`
	Personal      string `json:"personal,omitempty"`
	StackOverflow string `json:"stackoverflow,omitempty"`
}

// Experience represents a single employment experience
type Experience struct {
	Company     string    `json:"company"`
	Position    string    `json:"position"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Education represents an educational qualification
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Field       string `json:"field,omitempty"`
	GradYear    int    `json:"gradYear,omitempty"`
}

// GeoLocation represents geographical coordinates
type GeoLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// Location represents a full address with geo coordinates
type Location struct {
	FormattedAddress string      `json:"formattedAddress"`
	Country          string      `json:"country"`
	CountryCode      string      `json:"countryCode"`
	State            string      `json:"state"`
	StateCode        string      `json:"stateCode"`
	Geo              GeoLocation `json:"geo"`
}

// ResearchProfile contains comprehensive professional profile data from research
type ResearchProfile struct {
	ID                   string           `json:"id"`
	CanonicalURL         string           `json:"canonical"`
	FullName             string           `json:"fullName"`
	Headline             string           `json:"headline,omitempty"`
	Location             Location         `json:"location"`
	Industry             string           `json:"industry,omitempty"`
	ProfilePictureURL    string           `json:"pictureUrl,omitempty"`
	Social               SocialMediaLinks `json:"social"`
	CurrentExperiences   []Experience     `json:"currentExperiences"`
	PreviousExperiences  []Experience     `json:"previousExperiences"`
	Education            []Education      `json:"education"`
	Skills               []string         `json:"skills"`
	About                string           `json:"about,omitempty"`
	ExperienceInDays     int              `json:"experienceInDays"`
	HasMilitaryService   bool             `json:"hasMilitaryService"`
	HasSecurityClearance bool             `json:"hasSecurityClearance"`
}

// ResearchSubmissionSignal represents the signal sent when employment research is completed
type ResearchSubmissionSignal struct {
	Verified bool            `json:"verified"`
	Profile  ResearchProfile `json:"profile"`
}

// Config holds configuration for employment verification services
type Config struct {
	Keys      map[string]string
	Endpoints map[string]string
}

// VerificationRequest contains common fields for all verification requests
type VerificationRequest struct {
	EmployerName    string
	EmployerContact string
	CandidateName   string
}

// VerificationResult contains common fields for all verification results
type VerificationResult struct {
	Verified   bool
	VerifiedAt time.Time
	VerifiedBy string
	Notes      string
}

// EmploymentVerification contains employment verification information
type EmploymentVerification struct {
	CompanyName      string
	Position         string
	StartDate        string
	EndDate          string
	CurrentSalary    string
	ReasonForLeaving string
}

// VerificationSubmissionSignal represents the signal sent when employment verification is completed
type VerificationSubmissionSignal struct {
	Verified bool
	Profile  EmploymentVerification
}

// EmailParams represents parameters for sending an email
type EmailParams struct {
	To           string
	From         string
	Subject      string
	TemplateData map[string]interface{}
}

// EmailResponse represents the outcome of sending an email
type EmailResponse struct {
	Sent      bool
	MessageID string
	SentAt    time.Time
}

// Support email addresses
const (
	HiringManagerEmail     = "Hiring Manager <hiring@company.com>"
	HiringSupportEmail     = "Hiring Support <support@totaltalent.ai>"
	CandidateSupportEmail  = "Candidate <candidates@totaltalent.ai>"
	ResearcherSupportEmail = "Researcher <researcher@totaltalent.ai>"
)

// Request represents an email request
type Request struct {
	To           string
	From         string
	Subject      string
	TemplateData map[string]interface{}
}

// Result represents the outcome of sending an email
type Result struct {
	Sent      bool
	MessageID string
	SentAt    time.Time
}
