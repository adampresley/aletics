package services

import (
	"fmt"
	"time"

	"github.com/adampresley/aletics/internal/models"
	"gorm.io/gorm"
)

type ReportService struct {
	db *gorm.DB
}

type ReportServiceConfig struct {
	DB *gorm.DB
}

func NewReportService(config ReportServiceConfig) *ReportService {
	return &ReportService{
		db: config.DB,
	}
}

/*
GetViewsOverTime retrieves page view counts grouped by a specific time frame (day, hour).
This function is database-agnostic and supports both SQLite and PostgreSQL.
*/
func (s *ReportService) GetViewsOverTime(propertyID uint, start, end time.Time, timeframe string) ([]models.ViewsOverTimeItem, error) {
	var (
		err       error
		results   []models.ViewsOverTimeItem
		selectSQL string
		baseQuery *gorm.DB
	)

	switch s.db.Dialector.Name() {
	case "sqlite":
		switch timeframe {
		case "hourly":
			selectSQL = "strftime('%Y-%m-%d %H:00', created_at) as label, COUNT(*) as count"
		case "daily":
			selectSQL = "strftime('%Y-%m-%d', created_at) as label, COUNT(*) as count"
		default:
			return nil, fmt.Errorf("invalid timeframe for sqlite: %s", timeframe)
		}

	case "postgres":
		switch timeframe {
		case "hourly":
			selectSQL = "DATE_TRUNC('hour', created_at) as label, COUNT(*) as count"
		case "daily":
			selectSQL = "DATE_TRUNC('day', created_at)::date as label, COUNT(*) as count"
		default:
			return nil, fmt.Errorf("invalid timeframe for postgres: %s", timeframe)
		}

	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", s.db.Dialector.Name())
	}

	baseQuery = s.db.
		Model(&models.Event{}).
		Select(selectSQL).
		Where("property_id = ?", propertyID).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("label").
		Order("label ASC")

	if err = baseQuery.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetTopPaths returns the top 10 most viewed paths for a property within a given time range.
func (s *ReportService) GetTopPaths(propertyID uint, start, end time.Time) ([]models.TopPathItem, error) {
	var (
		err     error
		results []models.TopPathItem
	)

	err = s.db.
		Model(&models.Event{}).
		Select("path, COUNT(*) as count").
		Where("property_id = ?", propertyID).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("path").
		Order("count DESC").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetBrowserCounts returns the number of views per browser for a property within a given time range.
func (s *ReportService) GetBrowserCounts(propertyID uint, start, end time.Time) ([]models.BrowserCountItem, error) {
	var (
		err     error
		results []models.BrowserCountItem
	)

	err = s.db.
		Model(&models.Event{}).
		Select("browser, COUNT(*) as count").
		Where("property_id = ?", propertyID).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("browser").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
