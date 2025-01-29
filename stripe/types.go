package stripe

// Product represents a Stripe product
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

// Price represents a Stripe price
type Price struct {
	ID         string `json:"id"`
	ProductID  string `json:"product"`
	UnitAmount int64  `json:"unit_amount"`
	Currency   string `json:"currency"`
	Recurring  bool   `json:"recurring"`
}

// LineItem represents an item in a payment link
type LineItem struct {
	PriceID     string `json:"price"`
	Quantity    int64  `json:"quantity"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitAmount  int64  `json:"unit_amount"`
}

// PaymentLink represents a Stripe payment link
type PaymentLink struct {
	ID     string     `json:"id"`
	URL    string     `json:"url"`
	Active bool       `json:"active"`
	Items  []LineItem `json:"line_items"`
}

// AccountStatus represents the verification status of a Stripe account
type AccountStatus string

const (
	AccountStatusPending  AccountStatus = "pending"
	AccountStatusVerified AccountStatus = "verified"
	AccountStatusRejected AccountStatus = "rejected"
)

// Account represents a Stripe connected account
type Account struct {
	ID             string        `json:"id"`
	Email          string        `json:"email"`
	BusinessName   string        `json:"business_name"`
	Status         AccountStatus `json:"status"`
	PayoutsEnabled bool          `json:"payouts_enabled"`
}
