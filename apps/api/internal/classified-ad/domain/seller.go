package domain

// Seller is an internal entity representing the person selling a classified ad.
type Seller struct {
	email    Email
	pseudo   string
	password Password
}

// NewSeller validates and builds a Seller.
func NewSeller(email Email, pseudo string, password Password) (Seller, error) {
	if pseudo == "" {
		return Seller{}, ErrEmptyPseudo
	}
	if len(pseudo) > 30 {
		return Seller{}, ErrPseudoTooLong
	}
	return Seller{email: email, pseudo: pseudo, password: password}, nil
}

// Email returns the seller's email.
func (s Seller) Email() Email {
	return s.email
}

// Pseudo returns the seller's pseudo.
func (s Seller) Pseudo() string {
	return s.pseudo
}

// Password returns the seller's password.
func (s Seller) Password() Password {
	return s.password
}
