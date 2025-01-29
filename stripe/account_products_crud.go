package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/product"
)

// CreateAccountProductRequest contains the necessary information to create a product
type CreateAccountProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

// CreateAccountProduct creates a new product for a connected account
//
//encore:api method=POST path=/api/stripe/accounts/:accountID/products
func CreateAccountProduct(ctx context.Context, accountID string, req *CreateAccountProductRequest) (*Product, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	if req.Name == "" {
		return nil, errors.New("product name is required")
	}

	// Set up the params for creating a product
	params := &stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
		Active:      stripe.Bool(req.Active),
	}

	// Set the Stripe account to use
	params.SetStripeAccount(accountID)

	// Create the product
	p, err := product.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %v", err)
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Active:      p.Active,
	}, nil
}

// UpdateAccountProductRequest contains the fields that can be updated for a product
type UpdateAccountProductRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// UpdateAccountProduct updates an existing product for a connected account
//
//encore:api method=PUT path=/api/stripe/accounts/:accountID/products/:productID
func UpdateAccountProduct(ctx context.Context, accountID string, productID string, req *UpdateAccountProductRequest) (*Product, error) {
	if err := Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" || productID == "" {
		return nil, errors.New("account ID and product ID are required")
	}

	// Set up the params for updating a product
	params := &stripe.ProductParams{}

	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if req.Active != nil {
		params.Active = stripe.Bool(*req.Active)
	}

	// Set the Stripe account to use
	params.SetStripeAccount(accountID)

	// Update the product
	p, err := product.Update(productID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Active:      p.Active,
	}, nil
}

// DeleteAccountProduct deletes a product for a connected account
//
//encore:api method=DELETE path=/api/stripe/accounts/:accountID/products/:productID
func DeleteAccountProduct(ctx context.Context, accountID string, productID string) error {
	if err := Initialize(); err != nil {
		return fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" || productID == "" {
		return errors.New("account ID and product ID are required")
	}

	// Set up the params for deleting a product
	params := &stripe.ProductParams{}
	params.SetStripeAccount(accountID)

	// Delete the product
	_, err := product.Del(productID, params)
	if err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}

	return nil
}

// ActivateAccountProduct activates a product for a connected account
//
//encore:api method=POST path=/api/stripe/accounts/:accountID/products/:productID/activate
func ActivateAccountProduct(ctx context.Context, accountID string, productID string) error {
	if err := Initialize(); err != nil {
		return fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" || productID == "" {
		return errors.New("account ID and product ID are required")
	}

	// Set up the params for activating a product
	params := &stripe.ProductParams{
		Active: stripe.Bool(true),
	}
	params.SetStripeAccount(accountID)

	// Update the product to activate it
	_, err := product.Update(productID, params)
	if err != nil {
		return fmt.Errorf("failed to activate product: %v", err)
	}

	return nil
}

// DeactivateAccountProduct deactivates a product for a connected account
//
//encore:api method=POST path=/api/stripe/accounts/:accountID/products/:productID/deactivate
func DeactivateAccountProduct(ctx context.Context, accountID string, productID string) error {
	if err := Initialize(); err != nil {
		return fmt.Errorf("failed to initialize Stripe: %v", err)
	}

	if accountID == "" || productID == "" {
		return errors.New("account ID and product ID are required")
	}

	// Set up the params for deactivating a product
	params := &stripe.ProductParams{
		Active: stripe.Bool(false),
	}
	params.SetStripeAccount(accountID)

	// Update the product to deactivate it
	_, err := product.Update(productID, params)
	if err != nil {
		return fmt.Errorf("failed to deactivate product: %v", err)
	}

	return nil
}
