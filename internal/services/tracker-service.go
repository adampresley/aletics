package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/adampresley/aletics/internal/models"
	"gorm.io/gorm"
)

type TrackerService struct {
	db *gorm.DB
}

type TrackerServiceConfig struct {
	DB *gorm.DB
}

func NewTrackerService(config TrackerServiceConfig) *TrackerService {
	return &TrackerService{
		db: config.DB,
	}
}

func (s *TrackerService) TrackEvent(newEvent models.NewEvent) (*models.Event, error) {
	var (
		err       error
		property  models.Property
		originUrl *url.URL
	)

	if newEvent.Token == "" {
		return nil, fmt.Errorf("property 'token' is required")
	}

	if err = s.db.Where("token = ?", newEvent.Token).First(&property).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("property not found")
		}

		return nil, fmt.Errorf("error retrieving property by token: %w", err)
	}

	if !property.Active {
		return nil, fmt.Errorf("property is not active")
	}

	/*
	 * Validate the request origin against the property's domain. If
	 * the hostname from the request origina does not match the property's
	 * domain, return an error.
	 */
	if newEvent.Origin != "" {
		if originUrl, err = url.Parse(newEvent.Origin); err == nil {
			if property.Domain != originUrl.Hostname() {
				return nil, fmt.Errorf("request '%s' origin does not match property domain '%s'", originUrl.Hostname(), property.Domain)
			}
		}
	}

	queryString := newEvent.QueryString
	queryString = strings.TrimPrefix(queryString, "?")

	event := &models.Event{
		PropertyID:    property.ID,
		Path:          newEvent.Path,
		QueryString:   queryString,
		Browser:       newEvent.Browser,
		Country:       newEvent.Country,
		CountryCode:   newEvent.CountryCode,
		Continent:     newEvent.Continent,
		ContinentCode: newEvent.ContinentCode,
	}

	err = s.db.Create(event).Error
	return event, err
}
