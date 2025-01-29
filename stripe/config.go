package stripe

import (
	"github.com/stripe/stripe-go/v78"
)

// Initialize sets up the Stripe client with the provided API key
func Initialize() error {
	apiKey := secrets.StripeSecretKey
	if apiKey == "" {
		return &ErrMissingAPIKey
	}
	stripe.Key = apiKey
	return nil
}

// ErrMissingAPIKey is returned when the Stripe API key is not set
var ErrMissingAPIKey = stripe.Error{
	Code:   "config_error",
	Msg:    "Stripe API key is not set",
	Param:  "api_key",
	Type:   stripe.ErrorType("invalid_request_error"),
	DocURL: "https://stripe.com/docs/keys",
}
