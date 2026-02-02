package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"strings"

	"github.com/adampresley/aletics/internal/configuration"
	"github.com/adampresley/aletics/internal/handlers"
	"github.com/adampresley/mux"
	"github.com/adampresley/rendering"
	// "gorm.io/gorm"
)

var (
	Version string = "development"

	//go:embed app
	appFS embed.FS

	// db       *gorm.DB
	renderer rendering.TemplateRenderer

	dashboardHandler handlers.DashboardHandler
)

func main() {
	var (
		err error
	)

	config := configuration.LoadConfig()
	setupLogging(&config)
	shutdownCtx, stopApp := context.WithCancel(context.Background())

	if renderer, err = rendering.NewGoTemplateRenderer(appFS); err != nil {
		panic(err)
	}

	dashboardHandler = handlers.NewDashboardHandler(handlers.DashboardHandlerConfig{
		Config:   &config,
		Renderer: renderer,
	})

	routes := []mux.Route{
		{Path: "/", HandlerFunc: dashboardHandler.DashboardPage},
	}

	muxer := mux.Setup(
		config,
		routes,
		shutdownCtx,
		stopApp,

		mux.WithStaticContent("app", "/static/", appFS),
		mux.WithDebug(true),
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
