package main

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adampresley/aletics/internal/configuration"
	"github.com/adampresley/aletics/internal/handlers"
	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/mux"
	"github.com/adampresley/rendering"
	"github.com/adampresley/rester/clientoptions"
	"github.com/gorilla/sessions"
	"github.com/jellydator/ttlcache/v3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Version string = "development"

	//go:embed app
	appFS embed.FS

	db       *gorm.DB
	renderer rendering.TemplateRenderer
	store    *sessions.CookieStore

	dashboardHandler   *handlers.DashboardHandler
	propertyHandler    *handlers.PropertyHandler
	trackerHandler     *handlers.TrackerHandler
	userScriptsHandler *handlers.UserScriptsHandler
)

func main() {
	var (
		err     error
		dialect gorm.Dialector
	)

	config := configuration.LoadConfig()
	setupLogging(&config)
	shutdownCtx, stopApp := context.WithCancel(context.Background())

	/*
	 * Database
	 */
	if strings.HasPrefix(config.DSN, "file:") {
		dialect = sqlite.Open(config.DSN)
	} else if strings.HasPrefix(config.DSN, "postgres:") {
		dialect = postgres.Open(config.DSN)
	} else {
		panic("Unsupported database dialect")
	}

	if db, err = gorm.Open(dialect, &gorm.Config{}); err != nil {
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

	store = sessions.NewCookieStore([]byte(config.CookieSecret))

	/*
	 * Services
	 */
	ipCache := ttlcache.New(
		ttlcache.WithTTL[string, *models.CountryLookup](12 * time.Hour),
	)

	go ipCache.Start()

	propertyService := services.NewPropertyService(services.PropertyServiceConfig{
		DB: db,
	})

	trackerService := services.NewTrackerService(services.TrackerServiceConfig{
		DB: db,
	})

	reportService := services.NewReportService(services.ReportServiceConfig{
		DB: db,
	})

	restConfig := clientoptions.New(
		services.MaxmindBaseUrl,
		clientoptions.WithBasicAuth(config.MaxmindAccountID, config.MaxmindApiKey),
	)

	ipLookupService := services.NewIpLookupService(services.IpLookupServiceConfig{
		ApiAccountId: config.MaxmindAccountID,
		ApiKey:       config.MaxmindApiKey,
		RestConfig:   restConfig,
	})

	/*
	 * Handlers
	 */
	dashboardHandler = handlers.NewDashboardHandler(handlers.DashboardHandlerConfig{
		PropertyService: propertyService,
		ReportService:   reportService,
		Renderer:        renderer,
		ServerPassword:  config.ServerPassword,
		Store:           store,
	})

	propertyHandler = handlers.NewPropertyHandler(handlers.PropertyHandlerConfig{
		PropertyService: propertyService,
		Renderer:        renderer,
		TLD:             config.TLD,
	})

	trackerHandler = handlers.NewTrackerHandler(handlers.TrackerHandlerConfig{
		IpCache:         ipCache,
		IpLookupService: ipLookupService,
		TrackerService:  trackerService,
	})

	userScriptsHandler = handlers.NewUserScriptsHandler(handlers.UserScriptsHandlerConfig{
		FS: appFS,
	})

	routes := []mux.Route{
		{Path: "GET /aletics/v1/tracker.js", HandlerFunc: userScriptsHandler.TrackerScript, Middlewares: []mux.MiddlewareFunc{trackerCorsMiddleware}},
		{Path: "POST /aletics/v1/track", HandlerFunc: trackerHandler.TrackEvent, Middlewares: []mux.MiddlewareFunc{trackerCorsMiddleware}},

		{Path: "/", HandlerFunc: dashboardHandler.DashboardPage, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "GET /login", HandlerFunc: dashboardHandler.LoginPage},
		{Path: "POST /login", HandlerFunc: dashboardHandler.LoginAction},
		{Path: "GET /logout", HandlerFunc: dashboardHandler.LogoutAction},
		{Path: "GET /properties", HandlerFunc: propertyHandler.ManagePropertiesPage, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "GET /properties/create", HandlerFunc: propertyHandler.CreatePropertyPage, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "POST /properties/create", HandlerFunc: propertyHandler.CreatePropertyAction, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "GET /properties/edit/{id}", HandlerFunc: propertyHandler.EditPropertyPage, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "POST /properties/edit/{id}", HandlerFunc: propertyHandler.EditPropertyAction, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
		{Path: "DELETE /properties/delete/{id}", HandlerFunc: propertyHandler.DeleteProperty, Middlewares: []mux.MiddlewareFunc{authMiddleware}},
	}

	muxer := mux.Setup(
		&config,
		routes,
		shutdownCtx,
		stopApp,

		mux.WithStaticContent("app", "/static/", appFS),
		mux.WithDebug(Version == "development"),
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

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			session *sessions.Session
		)

		slog.Debug("authMiddleware", "request", r.URL.Path)

		if session, err = store.Get(r, "aletics_session"); err != nil {
			slog.Error("error retrieving session in authMiddleware", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["authenticated"]; !ok {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}
