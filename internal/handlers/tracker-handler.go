package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/aletics/internal/services"
	"github.com/adampresley/httphelpers/requests"
	"github.com/adampresley/httphelpers/responses"
	"github.com/jellydator/ttlcache/v3"
)

type TrackerHandler struct {
	ipCache         *ttlcache.Cache[string, *models.CountryLookup]
	ipLookupService *services.IpLookupService
	trackerService  *services.TrackerService
}

type TrackerHandlerConfig struct {
	IpCache         *ttlcache.Cache[string, *models.CountryLookup]
	IpLookupService *services.IpLookupService
	TrackerService  *services.TrackerService
}

func NewTrackerHandler(config TrackerHandlerConfig) *TrackerHandler {
	return &TrackerHandler{
		ipCache:         config.IpCache,
		ipLookupService: config.IpLookupService,
		trackerService:  config.TrackerService,
	}
}

func (h *TrackerHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		b             []byte
		newEvent      = models.NewEvent{}
		event         = &models.Event{}
		cacheItem     *ttlcache.Item[string, *models.CountryLookup]
		ok            bool
		countryName   string
		countryCode   string
		continentName string
		continentCode string
	)

	ip := services.GetIP(r)
	slog.Info("checking cache for IP", "ip", ip)

	cacheItem, ok = h.ipCache.GetOrSetFunc(ip, func() *models.CountryLookup {
		slog.Info("ip cache miss", "ip", ip)

		if slices.Contains([]string{"127.0.0.1", "localhost", "::1"}, ip) {
			slog.Warn("skipping ip lookup for local address", "ip", ip)
			return nil
		}

		newCountryInfo, err := h.ipLookupService.GetCountryInfo(ip)

		if err != nil {
			slog.Error("error retrieving country info for IP", "ip", ip, "error", err)
			return nil
		}

		return newCountryInfo
	})

	if ok {
		slog.Info("ip cache hit", "ip", ip)
	}

	if cacheItem.Value() != nil {
		ci := cacheItem.Value()

		if ci.Country != nil {
			if _, ok = ci.Country.Names["en"]; ok {
				countryName = ci.Country.Names["en"]
			} else if len(ci.Country.Names) > 0 {
				// assign first available key in names map as country name
				for _, v := range ci.Country.Names {
					countryName = v
					break
				}
			}

			if ci.Country.IsoCode != nil {
				countryCode = *ci.Country.IsoCode
			}
		}

		if ci.Continent != nil {
			if _, ok = ci.Continent.Names["en"]; ok {
				continentName = ci.Continent.Names["en"]
			} else if len(ci.Continent.Names) > 0 {
				for _, v := range ci.Continent.Names {
					continentName = v
					break
				}
			}

			if ci.Continent.Code != nil {
				continentCode = *ci.Continent.Code
			}
		}
	}

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

	newEvent.Country = countryName
	newEvent.CountryCode = countryCode
	newEvent.Continent = continentName
	newEvent.ContinentCode = continentCode
	newEvent.Origin = r.Header.Get("Origin")

	if event, err = h.trackerService.TrackEvent(newEvent); err != nil {
		slog.Error("error tracking tracker event", "error", err)
		responses.TextInternalServerError(w, "Error writing tracker event")
		return
	}

	slog.Info("tracked event", "id", event.ID, "path", event.Path, "browser", event.Browser)
	responses.TextOK(w, "ok")
}
