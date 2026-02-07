package handlers

import (
	"log/slog"
	"net/http"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/aletics/internal/viewdata"
	"github.com/adampresley/httphelpers/requests"
	"github.com/adampresley/rendering"
)

type PropertyHandler struct {
	propertyService *services.PropertyService
	renderer        rendering.TemplateRenderer
}

type PropertyHandlerConfig struct {
	PropertyService *services.PropertyService
	Renderer        rendering.TemplateRenderer
}

func NewPropertyHandler(config PropertyHandlerConfig) *PropertyHandler {
	return &PropertyHandler{
		propertyService: config.PropertyService,
		renderer:        config.Renderer,
	}
}

func (h *PropertyHandler) ManagePropertiesPage(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		pageName = "pages/properties/manage"
		viewData viewdata.ManageProperties
	)

	viewData = viewdata.ManageProperties{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
			JavascriptIncludes: []rendering.JavascriptInclude{
				{
					Type: "module",
					Src:  "/static/js/",
				},
			},
		},
		Properties: []models.Property{},
		Name:       requests.Get[string](r, "name"),
	}

	if viewData.Properties, err = h.propertyService.ListProperties(viewData.Name); err != nil {
		slog.Error("error getting properties list", "error", err)
		viewData.IsError = true
		viewData.Message = "There was a problem getting your list of properties."

		h.renderer.Render(pageName, viewData, w)
		return
	}

	h.renderer.Render(pageName, viewData, w)
}

func (h *PropertyHandler) CreatePropertyPage(w http.ResponseWriter, r *http.Request) {
	var (
		pageName = "pages/properties/create"
	)

	viewData := viewdata.CreateProperty{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Property: models.Property{
			Name:   requests.Get[string](r, "name"),
			Domain: requests.Get[string](r, "domain"),
			Token:  "",
			Active: true,
		},
	}

	h.renderer.Render(pageName, viewData, w)
}

func (h *PropertyHandler) CreatePropertyAction(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		pageName = "pages/properties/create"
	)

	viewData := viewdata.CreateProperty{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Property: models.Property{
			Name:   requests.Get[string](r, "name"),
			Domain: requests.Get[string](r, "domain"),
			Token:  "",
			Active: true,
		},
	}

	if _, err = h.propertyService.CreateProperty(viewData.Property.Name, viewData.Property.Domain); err != nil {
		slog.Error("error creating property", "name", viewData.Property.Name, "domain", viewData.Property.Domain, "error", err)
		viewData.IsError = true
		viewData.Message = "There was a problem creating your property."

		h.renderer.Render(pageName, viewData, w)
		return
	}

	// if viewData.IsHtmx {
	// 	w.Header().Set("HX-Redirect", "/properties")
	// 	w.WriteHeader(http.StatusOK)
	// } else {
	http.Redirect(w, r, "/properties", http.StatusSeeOther)
	// }
}

func (h *PropertyHandler) EditPropertyPage(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	pageName := "pages/properties/edit"

	viewData := viewdata.EditProperty{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Property: models.Property{},
	}

	id := requests.Get[uint](r, "id")

	if viewData.Property, err = h.propertyService.GetProperty(id); err != nil {
		slog.Error("error getting property", "id", id, "error", err)
		viewData.IsError = true
		viewData.Message = "There was a problem getting your property."

		h.renderer.Render(pageName, viewData, w)
		return
	}

	h.renderer.Render(pageName, viewData, w)
}

func (h *PropertyHandler) EditPropertyAction(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	pageName := "pages/properties/edit"

	viewData := viewdata.EditProperty{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Property: models.Property{
			Name:   requests.Get[string](r, "name"),
			Domain: requests.Get[string](r, "domain"),
			Active: requests.Get[bool](r, "active"),
		},
	}

	id := requests.Get[uint](r, "id")
	viewData.Property.ID = id

	if err = h.propertyService.UpdateProperty(id, viewData.Property); err != nil {
		slog.Error("error updating property", "id", id, "error", err)
		viewData.IsError = true
		viewData.Message = "There was a problem updating your property."

		h.renderer.Render(pageName, viewData, w)
		return
	}

	h.renderer.Render(pageName, viewData, w)
}

func (h *PropertyHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	id := requests.Get[uint](r, "id")

	if err = h.propertyService.DeleteProperty(id); err != nil {
		slog.Error("error deleting property", "error", err, "id", id)
		return
	}

	w.Header().Set("HX-Redirect", "/properties")
	w.WriteHeader(http.StatusOK)
}
