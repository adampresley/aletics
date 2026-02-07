package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model

	PropertyID uint     `json:"propertyId"`
	Property   Property `json:"-"`

	Path        string `json:"path"`
	QueryString string `json:"queryString"`
	Browser     string `json:"browser"`
	Country     string `json:"country"`
}

type NewEvent struct {
	Token  string `json:"token"`
	Origin string `json:"-"`

	Path        string `json:"path"`
	QueryString string `json:"queryString"`
	Browser     string `json:"browser"`
}
