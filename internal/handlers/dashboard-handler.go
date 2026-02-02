package handlers

import (
	"net/http"

	"github.com/adampresley/aletics/internal/configuration"
	"github.com/adampresley/rendering"
)

type DashboardHandlerConfig struct {
	Config   *configuration.Config
	Renderer rendering.TemplateRenderer
}

type DashboardHandler struct {
	config   *configuration.Config
	renderer rendering.TemplateRenderer
}

func NewDashboardHandler(config DashboardHandlerConfig) DashboardHandler {
	return DashboardHandler{
		config:   config.Config,
		renderer: config.Renderer,
	}
}

func (h DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	pageName := "pages/dashboard"
	h.renderer.Render(pageName, nil, w)
}
