package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/httphelpers/requests"
	"github.com/adampresley/httphelpers/responses"
)

type TrackerHandler struct {
	trackerService *services.TrackerService
}

type TrackerHandlerConfig struct {
	TrackerService *services.TrackerService
}

func NewTrackerHandler(config TrackerHandlerConfig) *TrackerHandler {
	return &TrackerHandler{
		trackerService: config.TrackerService,
	}
}

func (h *TrackerHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		b        []byte
		newEvent = models.NewEvent{}
		event    = &models.Event{}
	)

	if b, err = requests.Bytes(r); err != nil {
		slog.Error("error reading tracker event body", "error", err)
		responses.TextInternalServerError(w, "Error reading tracker event body")
		return
	}

	if err = json.Unmarshal(b, &newEvent); err != nil {
		slog.Error("error parsing tracker event body", "error", err)
		responses.TextInternalServerError(w, "Error parsing tracker event body")
		return
	}

	newEvent.Origin = r.Header.Get("Origin")

	if event, err = h.trackerService.TrackEvent(newEvent); err != nil {
		slog.Error("error tracking tracker event", "error", err)
		responses.TextInternalServerError(w, "Error writing tracker event")
		return
	}

	slog.Info("tracked event", "id", event.ID, "path", event.Path, "browser", event.Browser)
	responses.TextOK(w, "ok")
}
