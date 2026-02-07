package services

import (
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
	queryString := newEvent.QueryString
	queryString = strings.TrimPrefix(queryString, "?")

	event := &models.Event{
		Path:        newEvent.Path,
		QueryString: queryString,
		Browser:     newEvent.Browser,
		Country:     newEvent.Country,
	}

	err := s.db.Save(event).Error
	return event, err
}
