package http

import "time"

// PostClassifiedAdRequest is the request body for posting a classified ad
type PostClassifiedAdRequest struct {
	SellerId      string   `json:"sellerId"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	PriceAmount   int64    `json:"priceAmount"`
	PriceCurrency string   `json:"priceCurrency"`
	Category      string   `json:"category"`
	PhotoURLs     []string `json:"photoUrls"`
}

// PostClassifiedAdResponse is the response body after posting a classified ad
type PostClassifiedAdResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ClassifiedAdViewResponse represents a classified ad returned by the API
type ClassifiedAdViewResponse struct {
	Id            string    `json:"id"`
	SellerId      string    `json:"sellerId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	PriceAmount   int64     `json:"priceAmount"`
	PriceCurrency string    `json:"priceCurrency"`
	Category      string    `json:"category"`
	PhotoURLs     []string  `json:"photoUrls"`
	Status        string    `json:"status"`
	PostedAt      time.Time `json:"postedAt"`
}

// ErrorResponse is the standard error payload
type ErrorResponse struct {
	Error string `json:"error"`
}
