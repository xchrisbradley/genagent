package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/product"
)

// ListAccountProductsRequest contains parameters for listing products
type ListAccountProductsRequest struct {
	Limit         int64  `json:"limit,omitempty"`
	StartingAfter string `json:"starting_after,omitempty"`
}

// ListAccountProductsResponse contains the list of products and pagination info
type ListAccountProductsResponse struct {
	Products []Product `json:"products"`
	HasMore  bool      `json:"has_more"`
}

// ListAccountProducts lists all products for a connected account
//
//encore:api method=GET path=/api/stripe/accounts/:accountID/products
func ListAccountProducts(ctx context.Context, accountID string, req *ListAccountProductsRequest) (*ListAccountProductsResponse, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	// Set up the params for listing products
	params := &stripe.ProductListParams{
		ListParams: stripe.ListParams{},
	}

	// Set pagination parameters if provided
	if req.Limit > 0 {
		params.Limit = stripe.Int64(req.Limit)
	}
	if req.StartingAfter != "" {
		params.StartingAfter = stripe.String(req.StartingAfter)
	}

	// Set the Stripe account to use
	params.SetStripeAccount(accountID)

	// List products from Stripe
	i := product.List(params)

	// Collect the products
	var products []Product
	for i.Next() {
		p := i.Product()
		products = append(products, Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Active:      p.Active,
		})
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("failed to list products: %v", err)
	}

	return &ListAccountProductsResponse{
		Products: products,
		HasMore:  i.Meta().HasMore,
	}, nil
}
