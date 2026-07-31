package domain

// Price is a value object representing a monetary amount in cents.
type Price struct {
	amountInCents int64
}

// NewPrice validates and builds a Price from an amount in cents.
func NewPrice(cents int64) (Price, error) {
	if cents < 0 {
		return Price{}, ErrNegativePrice
	}
	return Price{amountInCents: cents}, nil
}

// AmountInCents returns the price amount in cents.
func (p Price) AmountInCents() int64 {
	return p.amountInCents
}
