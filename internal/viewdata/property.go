package viewdata

import (
	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/rendering"
)

type ManageProperties struct {
	rendering.BaseViewModel
	Properties []models.Property
	Name       string
}

type CreateProperty struct {
	rendering.BaseViewModel
	models.Property
}

type EditProperty struct {
	rendering.BaseViewModel
	models.Property
}
