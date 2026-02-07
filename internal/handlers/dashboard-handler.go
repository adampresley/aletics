package handlers

import (
	"net/http"

	"github.com/adampresley/aletics/internal/configuration"
	"github.com/adampresley/aletics/internal/viewdata"
	"github.com/adampresley/httphelpers/requests"
	"github.com/adampresley/httphelpers/responses"
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

	viewData := viewdata.Dashboard{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		ExampleData: "",
	}

	if r.URL.Path != "/" {
		responses.Text(w, http.StatusNotFound, "Not Found")
		return
	}

	h.renderer.Render(pageName, viewData, w)
}
