package handlers

import (
	"cmp"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/aletics/internal/viewdata"
	"github.com/adampresley/httphelpers/requests"
	"github.com/adampresley/rendering"
)

type DashboardHandler struct {
	propertyService *services.PropertyService
	reportService   *services.ReportService
	renderer        rendering.TemplateRenderer
}

type DashboardHandlerConfig struct {
	PropertyService *services.PropertyService
	ReportService   *services.ReportService
	Renderer        rendering.TemplateRenderer
}

func NewDashboardHandler(config DashboardHandlerConfig) *DashboardHandler {
	return &DashboardHandler{
		propertyService: config.PropertyService,
		reportService:   config.ReportService,
		renderer:        config.Renderer,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	var (
		err                   error
		pageName              = "pages/dashboard"
		properties            []models.Property
		selectedPropertyID    uint
		selectedTimeRange     string
		viewData              viewdata.Dashboard
		start, end            time.Time
		timeframe             string
		viewsOverTimeLabels   = make([]string, 0)
		viewsOverTimeData     = make([]int, 0)
		viewsOverTimeJSON     []byte
		viewsOverTimeDataJSON []byte
	)

	/*
	 * If we arrive here for a url that isn't /, it's a 404
	 */
	if r.URL.Path != "/" {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}

	/*
	 * Get all properties for the dropdown
	 */
	if properties, err = h.propertyService.ListProperties(""); err != nil {
		slog.Error("error getting properties list", "error", err)
	}

	/*
	 * Get filter values from the request, with defaults
	 */
	selectedPropertyID = requests.Get[uint](r, "property_id")
	selectedTimeRange = cmp.Or(requests.Get[string](r, "time_range"), "7d")

	if selectedPropertyID == 0 && len(properties) > 0 {
		selectedPropertyID = properties[0].ID
	}

	start, end, timeframe = calculateDateRange(selectedTimeRange)

	viewData = viewdata.Dashboard{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Properties:         properties,
		SelectedPropertyID: selectedPropertyID,
		SelectedTimeRange:  selectedTimeRange,
	}

	/*
	 * If we have a property, get the report data
	 */
	if selectedPropertyID > 0 {
		if viewData.ViewsOverTime, err = h.reportService.GetViewsOverTime(selectedPropertyID, start, end, timeframe); err != nil {
			slog.Error("error getting views over time", "error", err)
		}

		if viewData.TopPaths, err = h.reportService.GetTopPaths(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting top paths", "error", err)
		}

		if viewData.BrowserCounts, err = h.reportService.GetBrowserCounts(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting browser counts", "error", err)
		}
	}

	/*
	 * Prepare data for Chart.js
	 */
	for _, item := range viewData.ViewsOverTime {
		viewsOverTimeLabels = append(viewsOverTimeLabels, item.Label)
		viewsOverTimeData = append(viewsOverTimeData, item.Count)
	}

	if viewsOverTimeJSON, err = json.Marshal(viewsOverTimeLabels); err == nil {
		viewData.ViewsOverTimeLabelsJSON = template.JS(viewsOverTimeJSON)
	}

	if viewsOverTimeDataJSON, err = json.Marshal(viewsOverTimeData); err == nil {
		viewData.ViewsOverTimeDataJSON = template.JS(viewsOverTimeDataJSON)
	}

	h.renderer.Render(pageName, viewData, w)
}

func calculateDateRange(timeRange string) (time.Time, time.Time, string) {
	end := time.Now()
	var start time.Time
	timeframe := "daily"

	switch timeRange {
	case "24h":
		start = end.Add(-24 * time.Hour)
		timeframe = "hourly"
	case "1d":
		yesterday := end.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		end = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())
		timeframe = "hourly"
	case "30d":
		start = end.AddDate(0, -1, 0)
	case "6m":
		start = end.AddDate(0, -6, 0)
	case "7d":
	default:
		start = end.AddDate(0, 0, -7)
	}

	return start, end, timeframe
}
