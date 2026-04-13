package viewdata

import (
	"html/template"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/rendering"
)

type Dashboard struct {
	rendering.BaseViewModel

	// Fields for filter controls
	Properties         []models.Property
	SelectedPropertyID uint
	SelectedTimeRange  string

	// Report data
	ViewsOverTime    []models.ViewsOverTimeItem
	TopPaths         []models.TopPathItem
	BrowserCounts    []models.BrowserCountItem
	CountryCounts    []models.CountryCountItem
	JourneyEdges     []models.JourneyEdge
	TopJourneyPaths  []models.TopJourneyPath

	// Data formatted for Chart.js / Sankey, must be template.JS to be safe
	ViewsOverTimeLabelsJSON template.JS
	ViewsOverTimeDataJSON   template.JS
	JourneyEdgesJSON        template.JS
}

type Login struct {
	rendering.BaseViewModel
	Password string
}
