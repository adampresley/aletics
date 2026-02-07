package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model

	Path        string `json:"path"`
	QueryString string `json:"queryString"`
	Browser     string `json:"browser"`
	Country     string `json:"country"`
}

type NewEvent struct {
	Path        string `json:"path"`
	QueryString string `json:"queryString"`
	Browser     string `json:"browser"`
	Country     string `json:"country"`
}
