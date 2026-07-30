package domain

import "errors"

var (
	ErrEmptyTitle           = errors.New("title must not be empty")
	ErrTitleTooLong         = errors.New("title must not exceed 100 characters")
	ErrEmptyDescription     = errors.New("description must not be empty")
	ErrDescriptionTooLong   = errors.New("description must not exceed 4000 characters")
	ErrNegativePrice        = errors.New("price must not be negative")
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrEmptyPseudo          = errors.New("pseudo must not be empty")
	ErrPseudoTooLong        = errors.New("pseudo must not exceed 30 characters")
	ErrPasswordTooShort     = errors.New("password must be at least 8 characters long")
	ErrInvalidZipCode       = errors.New("zip code must be exactly 5 digits")
	ErrEmptyCityName        = errors.New("city name must not be empty")
	ErrInvalidCategory      = errors.New("invalid category")
	ErrInvalidDeleteReason  = errors.New("invalid delete reason")
	ErrTooManyImages        = errors.New("a classified ad cannot have more than 10 images")
	ErrEmptyImageURL        = errors.New("image url must not be empty")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAdNotAvailable       = errors.New("classified ad is not available")
	ErrClassifiedAdNotFound = errors.New("classified ad not found")
	ErrEmptyOfferMessage    = errors.New("offer message must not be empty")
	ErrOfferMessageTooLong  = errors.New("offer message must not exceed 1000 characters")
	ErrNegativeOfferAmount  = errors.New("offer amount must not be negative")
)
