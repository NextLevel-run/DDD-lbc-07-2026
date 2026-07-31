package domain

import "regexp"

var zipCodeRegex = regexp.MustCompile(`^\d{5}$`)

// Location is an internal entity representing where a classified ad is located.
type Location struct {
	zipCode  string
	cityName string
}

// NewLocation validates and builds a Location.
func NewLocation(zipCode, cityName string) (Location, error) {
	if !zipCodeRegex.MatchString(zipCode) {
		return Location{}, ErrInvalidZipCode
	}
	if cityName == "" {
		return Location{}, ErrEmptyCityName
	}
	return Location{zipCode: zipCode, cityName: cityName}, nil
}

// ZipCode returns the location's zip code.
func (l Location) ZipCode() string {
	return l.zipCode
}

// CityName returns the location's city name.
func (l Location) CityName() string {
	return l.cityName
}
