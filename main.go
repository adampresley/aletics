package main

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/adampresley/aletics/internal/configuration"
	"github.com/adampresley/aletics/internal/handlers"
	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/mux"
	"github.com/adampresley/rendering"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Version string = "development"

	//go:embed app
	appFS embed.FS

	db       *gorm.DB
	renderer rendering.TemplateRenderer

	dashboardHandler   *handlers.DashboardHandler
	propertyHandler    *handlers.PropertyHandler
	trackerHandler     *handlers.TrackerHandler
	userScriptsHandler *handlers.UserScriptsHandler
)

func main() {
	var (
		err error
	)

	config := configuration.LoadConfig()
	setupLogging(&config)
	shutdownCtx, stopApp := context.WithCancel(context.Background())

	/*
	 * Database
	 */
	if db, err = gorm.Open(sqlite.Open(config.DSN), &gorm.Config{}); err != nil {
		slog.Error("error connecting to database", "error", err)
		os.Exit(1)
	}

	slog.Info("Database connection established. Running migrations...")

	db.AutoMigrate(
		&models.Property{}, &models.Event{},
	)

	if renderer, err = rendering.NewGoTemplateRenderer(appFS); err != nil {
		panic(err)
	}

	/*
	 * Services
	 */
	propertyService := services.NewPropertyService(services.PropertyServiceConfig{
		DB: db,
	})

	trackerService := services.NewTrackerService(services.TrackerServiceConfig{
		DB: db,
	})
	reportService := services.NewReportService(services.ReportServiceConfig{
		DB: db,
	})

	/*
	 * Handlers
	 */
	dashboardHandler = handlers.NewDashboardHandler(handlers.DashboardHandlerConfig{
		PropertyService: propertyService,
		ReportService:   reportService,
		Renderer:        renderer,
	})

	propertyHandler = handlers.NewPropertyHandler(handlers.PropertyHandlerConfig{
		PropertyService: propertyService,
		Renderer:        renderer,
	})

	trackerHandler = handlers.NewTrackerHandler(handlers.TrackerHandlerConfig{
		TrackerService: trackerService,
	})

	userScriptsHandler = handlers.NewUserScriptsHandler(handlers.UserScriptsHandlerConfig{
		FS: appFS,
	})

	routes := []mux.Route{
		{Path: "/", HandlerFunc: dashboardHandler.DashboardPage},

		{Path: "GET /aletics/v1/tracker.js", HandlerFunc: userScriptsHandler.TrackerScript, Middlewares: []mux.MiddlewareFunc{trackerCorsMiddleware}},
		{Path: "POST /aletics/v1/track", HandlerFunc: trackerHandler.TrackEvent, Middlewares: []mux.MiddlewareFunc{trackerCorsMiddleware}},

		{Path: "GET /properties", HandlerFunc: propertyHandler.ManagePropertiesPage},
		{Path: "GET /properties/create", HandlerFunc: propertyHandler.CreatePropertyPage},
		{Path: "POST /properties/create", HandlerFunc: propertyHandler.CreatePropertyAction},
		{Path: "GET /properties/edit/{id}", HandlerFunc: propertyHandler.EditPropertyPage},
		{Path: "POST /properties/edit/{id}", HandlerFunc: propertyHandler.EditPropertyAction},
		{Path: "DELETE /properties/delete/{id}", HandlerFunc: propertyHandler.DeleteProperty},
	}

	muxer := mux.Setup(
		&config,
		routes,
		shutdownCtx,
		stopApp,

		mux.WithStaticContent("app", "/static/", appFS),
		mux.WithDebug(true),
		mux.WithMiddlewares(
			func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					slog.Info("request", "method", r.Method, "path", r.URL.Path)
					h.ServeHTTP(w, r)
				})
			},
		),
	)

	muxer.Start()
}

func setupLogging(config *configuration.Config) {
	var (
		logger *slog.Logger
	)

	level := slog.LevelInfo

	switch strings.ToLower(config.LogLevel) {
	case "debug":
		level = slog.LevelDebug

	case "error":
		level = slog.LevelError

	default:
		level = slog.LevelInfo
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}).WithAttrs([]slog.Attr{
		slog.String("version", Version),
	})

	logger = slog.New(h)
	slog.SetDefault(logger)
}

func trackerCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content, Content-Type, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		next.ServeHTTP(w, r)
	})
}
