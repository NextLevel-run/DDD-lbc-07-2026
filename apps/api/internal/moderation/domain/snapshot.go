package domain

// ClassifiedAdSnapshot is a value object capturing the full content of a
// classified ad at submission or edition time. It is a pure data carrier
// (copied from public integration events, no invariants of its own), hence
// the exported fields.
type ClassifiedAdSnapshot struct {
	Title        string
	Description  string
	PriceInCents int64
	ImageURLs    []string
	Category     string
	ZipCode      string
	CityName     string
	SellerEmail  string
	SellerPseudo string
}
