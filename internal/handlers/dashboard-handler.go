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
	"github.com/gorilla/sessions"
)

type DashboardHandler struct {
	propertyService *services.PropertyService
	reportService   *services.ReportService
	renderer        rendering.TemplateRenderer
	serverPassword  string
	store           *sessions.CookieStore
}

type DashboardHandlerConfig struct {
	PropertyService *services.PropertyService
	ReportService   *services.ReportService
	Renderer        rendering.TemplateRenderer
	ServerPassword  string
	Store           *sessions.CookieStore
}

func NewDashboardHandler(config DashboardHandlerConfig) *DashboardHandler {
	return &DashboardHandler{
		propertyService: config.PropertyService,
		reportService:   config.ReportService,
		renderer:        config.Renderer,
		serverPassword:  config.ServerPassword,
		store:           config.Store,
	}
}

func (h *DashboardHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	pageName := "pages/login"

	viewData := viewdata.Login{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Password: "",
	}

	h.renderer.Render(pageName, viewData, w)
}

func (h *DashboardHandler) LoginAction(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		session *sessions.Session
	)

	pageName := "pages/login"

	viewData := viewdata.Login{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		Password: requests.Get[string](r, "password"),
	}

	if viewData.Password != h.serverPassword {
		ip := services.GetIP(r)
		slog.Error("invalid loginn attempt", "ip", ip)

		viewData.IsError = true
		viewData.Message = "Invalid password"

		h.renderer.Render(pageName, viewData, w)
		return
	}

	if session, err = h.store.Get(r, "aletics_session"); err != nil {
		slog.Error("error getting session", "error", err)

		viewData.IsError = true
		viewData.Message = "An unexpected error occurred while validating your password. Please try again"

		h.renderer.Render(pageName, viewData, w)
		return
	}

	session.Values["authenticated"] = true

	if err = session.Save(r, w); err != nil {
		slog.Error("error saving session", "error", err)

		viewData.IsError = true
		viewData.Message = "An unexpected error occurred while validating your password. Please try again"

		h.renderer.Render(pageName, viewData, w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *DashboardHandler) LogoutAction(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		session *sessions.Session
	)

	if session, err = h.store.Get(r, "aletics_session"); err != nil {
		slog.Error("error getting session", "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session.Options.MaxAge = -1

	if err = session.Save(r, w); err != nil {
		slog.Error("error saving session", "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
		journeyEdgesJSON      []byte
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

		viewData.ViewsOverTime = fillMissingDataPoints(viewData.ViewsOverTime, start, end, timeframe)

		if viewData.TopPaths, err = h.reportService.GetTopPaths(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting top paths", "error", err)
		}

		if viewData.BrowserCounts, err = h.reportService.GetBrowserCounts(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting browser counts", "error", err)
		}

		if viewData.CountryCounts, err = h.reportService.GetCountryCounts(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting country counts", "error", err)
		}

		if viewData.JourneyEdges, err = h.reportService.GetJourneyEdges(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting journey edges", "error", err)
		}

		if viewData.TopJourneyPaths, err = h.reportService.GetTopJourneyPaths(selectedPropertyID, start, end); err != nil {
			slog.Error("error getting top journey paths", "error", err)
		}
	}

	/*
	 * Prepare data for Chart.js
	 */
	for _, item := range viewData.ViewsOverTime {
		viewsOverTimeLabels = append(viewsOverTimeLabels, formatChartLabel(item.Label, timeframe))
		viewsOverTimeData = append(viewsOverTimeData, item.Count)
	}

	if viewsOverTimeJSON, err = json.Marshal(viewsOverTimeLabels); err == nil {
		viewData.ViewsOverTimeLabelsJSON = template.JS(viewsOverTimeJSON)
	}

	if viewsOverTimeDataJSON, err = json.Marshal(viewsOverTimeData); err == nil {
		viewData.ViewsOverTimeDataJSON = template.JS(viewsOverTimeDataJSON)
	}

	if journeyEdgesJSON, err = json.Marshal(viewData.JourneyEdges); err == nil {
		viewData.JourneyEdgesJSON = template.JS(journeyEdgesJSON)
	}

	h.renderer.Render(pageName, viewData, w)
}

func calculateDateRange(timeRange string) (time.Time, time.Time, string) {
	var (
		end       time.Time
		start     time.Time
		timeframe string
		today     time.Time
		yesterday time.Time
	)

	end = time.Now()
	timeframe = "daily"
	today = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	switch timeRange {
	case "24h":
		start = end.Add(-24 * time.Hour)
		timeframe = "hourly"
	case "1d":
		yesterday = today.AddDate(0, 0, -1)
		start = yesterday
		end = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location())
		timeframe = "hourly"
	case "30d":
		start = today.AddDate(0, 0, -29)
	case "6m":
		start = today.AddDate(0, -6, 0)
	default: // includes "7d"
		start = today.AddDate(0, 0, -6)
	}

	return start, end, timeframe
}

func parseLabelTime(label string) (time.Time, bool) {
	var (
		formats = []string{
			time.RFC3339,
			"2006-01-02T15:04:05+00:00",
			"2006-01-02 15:04:05+00:00",
			"2006-01-02 15:04:05+00",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		t   time.Time
		err error
	)

	for _, format := range formats {
		if t, err = time.Parse(format, label); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func formatChartLabel(label, timeframe string) string {
	var (
		t  time.Time
		ok bool
	)

	t, ok = parseLabelTime(label)
	if !ok {
		return label
	}

	if timeframe == "hourly" {
		return t.Format("Jan 2 15:04")
	}

	return t.Format("Jan 2")
}

func fillMissingDataPoints(items []models.ViewsOverTimeItem, start, end time.Time, timeframe string) []models.ViewsOverTimeItem {
	var (
		existing map[string]int
		result   []models.ViewsOverTimeItem
		startUTC time.Time
		endUTC   time.Time
		endDay   time.Time
		current  time.Time
		key      string
		t        time.Time
		ok       bool
	)

	existing = make(map[string]int)
	result = make([]models.ViewsOverTimeItem, 0)
	startUTC = start.UTC()
	endUTC = end.UTC()

	for _, item := range items {
		t, ok = parseLabelTime(item.Label)
		if !ok {
			continue
		}

		if timeframe == "hourly" {
			key = t.UTC().Format("2006-01-02 15:00")
		} else {
			key = t.UTC().Format("2006-01-02")
		}

		existing[key] = item.Count
	}

	if timeframe == "hourly" {
		current = time.Date(startUTC.Year(), startUTC.Month(), startUTC.Day(), startUTC.Hour(), 0, 0, 0, time.UTC)
		for !current.After(endUTC) {
			key = current.Format("2006-01-02 15:00")
			result = append(result, models.ViewsOverTimeItem{Label: key, Count: existing[key]})
			current = current.Add(time.Hour)
		}
	} else {
		current = time.Date(startUTC.Year(), startUTC.Month(), startUTC.Day(), 0, 0, 0, 0, time.UTC)
		endDay = time.Date(endUTC.Year(), endUTC.Month(), endUTC.Day(), 0, 0, 0, 0, time.UTC)
		for !current.After(endDay) {
			key = current.Format("2006-01-02")
			result = append(result, models.ViewsOverTimeItem{Label: key, Count: existing[key]})
			current = current.AddDate(0, 0, 1)
		}
	}

	return result
}
