package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/product"
)

// CreateProductRequest contains the necessary information to create a Stripe product
type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateProduct creates a new Stripe product
//
//encore:api method=POST path=/api/stripe/products
func CreateProduct(ctx context.Context, req *CreateProductRequest) (*Product, error) {
	if req.Name == "" {
		return nil, errors.New("product name is required")
	}

	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	// Create the product params
	params := &stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
		Active:      stripe.Bool(true),
	}

	// Create the product
	p, err := product.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe product: %v", err)
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Active:      p.Active,
	}, nil
}
