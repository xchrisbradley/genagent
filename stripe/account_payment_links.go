package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentlink"
	"github.com/stripe/stripe-go/v78/price"
)

// CreateAccountPaymentLinkRequest contains the necessary information to create a payment link
type CreateAccountPaymentLinkRequest struct {
	UnitAmount int64  `json:"unit_amount"`
	Currency   string `json:"currency,omitempty"`
}

// CreateAccountPaymentLink creates a new payment link for a product in a connected account
//
//encore:api method=POST path=/api/stripe/accounts/:accountID/products/:productID/payment-links
func CreateAccountPaymentLink(ctx context.Context, accountID string, productID string, req *CreateAccountPaymentLinkRequest) (*PaymentLink, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" || productID == "" {
		return nil, errors.New("account ID and product ID are required")
	}

	if req.UnitAmount <= 0 {
		return nil, errors.New("unit amount must be greater than 0")
	}

	if req.UnitAmount < 50 {
		return nil, errors.New("unit amount must be at least $0.50 USD")
	}

	currency := "usd"
	if req.Currency != "" {
		currency = req.Currency
	}

	priceParams := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(req.UnitAmount),
		Currency:   stripe.String(currency),
	}
	priceParams.SetStripeAccount(accountID)

	p, err := price.New(priceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %v", err)
	}

	// Calculate 3% of the unit amount
	// Since Stripe amounts are in cents, we need to calculate accordingly
	// For example, if unit amount is 2000 (representing $20.00):
	// 2000 * 0.03 = 60 (representing $0.60)
	applicationFeeAmount := int64(float64(req.UnitAmount) * 0.03)

	params := &stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price:    stripe.String(p.ID),
				Quantity: stripe.Int64(1),
			},
		},
		AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
			Type: stripe.String(string(stripe.PaymentLinkAfterCompletionTypeRedirect)),
			Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
				URL: stripe.String("https://your-success-url.com"),
			},
		},
		ApplicationFeeAmount: stripe.Int64(applicationFeeAmount),
	}

	params.PaymentIntentData = &stripe.PaymentLinkPaymentIntentDataParams{
		CaptureMethod:    stripe.String(string(stripe.PaymentIntentCaptureMethodAutomatic)),
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	params.SetStripeAccount(accountID)

	pl, err := paymentlink.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment link: %v", err)
	}

	return &PaymentLink{
		ID:     pl.ID,
		URL:    pl.URL,
		Active: pl.Active,
	}, nil
}
