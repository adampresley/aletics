package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/adampresley/httphelpers/responses"
)

type UserScriptsHandler struct {
	fs fs.FS
}

type UserScriptsHandlerConfig struct {
	FS fs.FS
}

func NewUserScriptsHandler(config UserScriptsHandlerConfig) *UserScriptsHandler {
	return &UserScriptsHandler{
		fs: config.FS,
	}
}

func (h *UserScriptsHandler) TrackerScript(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		b   []byte
	)

	if b, err = fs.ReadFile(h.fs, "app/pages/user-scripts/tracker.js"); err != nil {
		slog.Error("error reading tracker script", "error", err)
		responses.TextInternalServerError(w, "Error reading tracker script")
		return
	}

	responses.Bytes(w, http.StatusOK, "application/javascript", b)
}
