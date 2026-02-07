package models

import "gorm.io/gorm"

type Property struct {
	gorm.Model

	Name   string
	Domain string
	Token  string `gorm:"unique"`
	Active bool
}
