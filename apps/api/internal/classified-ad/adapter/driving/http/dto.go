package httpadapter

import "time"

// Request DTOs

// SubmitClassifiedAdRequest is the request body for POST /classified-ads.
type SubmitClassifiedAdRequest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	PriceInCents   int64    `json:"priceInCents"`
	SellerEmail    string   `json:"sellerEmail"`
	SellerPseudo   string   `json:"sellerPseudo"`
	SellerPassword string   `json:"sellerPassword"`
	ImageURLs      []string `json:"imageUrls"`
	Category       string   `json:"category"`
	ZipCode        string   `json:"zipCode"`
	CityName       string   `json:"cityName"`
}

// MakeOfferRequest is the request body for POST /classified-ads/{id}/offers.
type MakeOfferRequest struct {
	BuyerEmail    string `json:"buyerEmail"`
	BuyerPseudo   string `json:"buyerPseudo"`
	AmountInCents int64  `json:"amountInCents"`
	Message       string `json:"message"`
}

// DeleteClassifiedAdRequest is the request body for DELETE /classified-ads/{id}.
type DeleteClassifiedAdRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Reason   string `json:"reason"`
}

// EditClassifiedAdRequest is the request body for PUT /classified-ads/{id}.
// The seller authenticates with their email and password, like for deletion.
type EditClassifiedAdRequest struct {
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	PriceInCents int64    `json:"priceInCents"`
	ImageURLs    []string `json:"imageUrls"`
	Category     string   `json:"category"`
	ZipCode      string   `json:"zipCode"`
	CityName     string   `json:"cityName"`
}

// Response DTOs

// SubmitClassifiedAdResponse is the response body for a successful submission.
type SubmitClassifiedAdResponse struct {
	ID string `json:"id"`
}

// ClassifiedAdListItemResponse is a single item in a search result list.
type ClassifiedAdListItemResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	PriceInCents   int64     `json:"priceInCents"`
	Category       string    `json:"category"`
	CityName       string    `json:"cityName"`
	ZipCode        string    `json:"zipCode"`
	FirstImageURL  string    `json:"firstImageUrl"`
	SubmissionDate time.Time `json:"submissionDate"`
}

// SearchClassifiedAdsResponse is the response body for the search endpoint.
type SearchClassifiedAdsResponse struct {
	Items []ClassifiedAdListItemResponse `json:"items"`
}

// ClassifiedAdViewResponse is the response body for the detail endpoint.
type ClassifiedAdViewResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	PriceInCents   int64     `json:"priceInCents"`
	Category       string    `json:"category"`
	SellerPseudo   string    `json:"sellerPseudo"`
	ImageURLs      []string  `json:"imageUrls"`
	ZipCode        string    `json:"zipCode"`
	CityName       string    `json:"cityName"`
	SubmissionDate time.Time `json:"submissionDate"`
}

// ErrorResponse is the standard error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}
