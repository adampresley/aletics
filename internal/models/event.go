package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model

	PropertyID uint     `json:"propertyId"`
	Property   Property `json:"-"`

	Path          string `json:"path"`
	QueryString   string `json:"queryString"`
	Browser       string `json:"browser"`
	Country       string `json:"country"`
	CountryCode   string `json:"countryCode"`
	Continent     string `json:"continent"`
	ContinentCode string `json:"continentCode"`
}

type NewEvent struct {
	Token  string `json:"token"`
	Origin string `json:"-"`

	Path          string `json:"path"`
	QueryString   string `json:"queryString"`
	Browser       string `json:"browser"`
	Country       string `json:"-"`
	CountryCode   string `json:"-"`
	Continent     string `json:"-"`
	ContinentCode string `json:"-"`
}
