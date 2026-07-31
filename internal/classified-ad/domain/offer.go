package domain

// ValidateOfferMessage validates a buyer's offer message.
func ValidateOfferMessage(message string) error {
	if message == "" {
		return ErrEmptyOfferMessage
	}
	if len(message) > 1000 {
		return ErrOfferMessageTooLong
	}
	return nil
}

// ValidateOfferAmount validates a buyer's offer amount, in cents.
func ValidateOfferAmount(amountInCents int64) error {
	if amountInCents < 0 {
		return ErrNegativeOfferAmount
	}
	return nil
}
