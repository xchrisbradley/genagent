package stripe

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/account"
	"github.com/stripe/stripe-go/v78/accountlink"
)

// CreateAccountRequest contains the necessary information to create a Stripe connected account
type CreateAccountRequest struct {
	Email        string `json:"email"`
	BusinessName string `json:"business_name"`
	Country      string `json:"country"` // Two-letter ISO country code
}

// CreateAccount creates a new Stripe connected account
//
//encore:api method=POST path=/api/stripe/accounts
func CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if req.Email == "" || req.BusinessName == "" || req.Country == "" {
		return nil, errors.New("email, business name, and country are required")
	}

	params := &stripe.AccountParams{
		Type:         stripe.String("express"),
		Email:        stripe.String(req.Email),
		BusinessType: stripe.String("company"),
		Company: &stripe.AccountCompanyParams{
			Name: stripe.String(req.BusinessName),
		},
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		Country: stripe.String(req.Country),
	}

	acct, err := account.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe account: %v", err)
	}

	status := AccountStatusPending
	if acct.PayoutsEnabled {
		status = AccountStatusVerified
	}

	return &Account{
		ID:             acct.ID,
		Email:          acct.Email,
		BusinessName:   acct.Company.Name,
		Status:         status,
		PayoutsEnabled: acct.PayoutsEnabled,
	}, nil
}

// GetAccountLinkRequest contains parameters for generating an account link
type GetAccountLinkRequest struct {
	ReturnURL  string `json:"return_url"`
	RefreshURL string `json:"refresh_url"`
}

// AccountLink represents a URL that users can visit to onboard or update their Stripe connected account
type AccountLink struct {
	URL string `json:"url"`
}

// GetAccountLink generates a Stripe account link for onboarding or updating account details
//
//encore:api method=POST path=/api/stripe/accounts/:id/link
func GetAccountLink(ctx context.Context, id string, req *GetAccountLinkRequest) (*AccountLink, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if id == "" {
		return nil, errors.New("account ID is required")
	}

	if req.ReturnURL == "" || req.RefreshURL == "" {
		return nil, errors.New("return URL and refresh URL are required")
	}

	// Validate URLs
	if _, err := url.Parse(req.ReturnURL); err != nil {
		return nil, fmt.Errorf("invalid return URL: %v", err)
	}
	if _, err := url.Parse(req.RefreshURL); err != nil {
		return nil, fmt.Errorf("invalid refresh URL: %v", err)
	}

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(id),
		ReturnURL:  stripe.String(req.ReturnURL),
		RefreshURL: stripe.String(req.RefreshURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %v", err)
	}

	return &AccountLink{URL: link.URL}, nil
}

// GetAccount retrieves a Stripe connected account's details
//
//encore:api method=GET path=/api/stripe/accounts/:id
func GetAccount(ctx context.Context, id string) (*Account, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if id == "" {
		return nil, errors.New("account ID is required")
	}

	acct, err := account.GetByID(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Stripe account: %v", err)
	}

	status := AccountStatusPending
	if acct.PayoutsEnabled {
		status = AccountStatusVerified
	}

	return &Account{
		ID:             acct.ID,
		Email:          acct.Email,
		BusinessName:   acct.BusinessProfile.Name,
		Status:         status,
		PayoutsEnabled: acct.PayoutsEnabled,
	}, nil
}
