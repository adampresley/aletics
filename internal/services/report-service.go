package services

import (
	"fmt"
	"sort"
	"strings"
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

// GetCountryCounts returns the number of views per country for a property within a given time range.
func (s *ReportService) GetCountryCounts(propertyID uint, start, end time.Time) ([]models.CountryCountItem, error) {
	var (
		err     error
		results []models.CountryCountItem
	)

	err = s.db.
		Model(&models.Event{}).
		Select("country, COUNT(*) as count").
		Where("property_id = ?", propertyID).
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("country").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetJourneyEdges returns direct page-to-page transitions for sessions within a given time range.
// It uses a window function CTE to find consecutive page pairs within each session.
func (s *ReportService) GetJourneyEdges(propertyID uint, start, end time.Time) ([]models.JourneyEdge, error) {
	var results []models.JourneyEdge

	cteSQL := `
		WITH ordered AS (
			SELECT
				session_id,
				path,
				ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at) AS rn
			FROM events
			WHERE property_id = ?
				AND session_id != ''
				AND deleted_at IS NULL
				AND created_at BETWEEN ? AND ?
		)
		SELECT a.path AS source, b.path AS target, COUNT(*) AS value
		FROM ordered a
		JOIN ordered b ON a.session_id = b.session_id AND b.rn = a.rn + 1
		GROUP BY a.path, b.path
		ORDER BY value DESC
		LIMIT 50
	`

	if err := s.db.Raw(cteSQL, propertyID, start, end).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetTopJourneyPaths returns the most common multi-page session sequences within a given time range.
func (s *ReportService) GetTopJourneyPaths(propertyID uint, start, end time.Time) ([]models.TopJourneyPath, error) {
	type sessionEvent struct {
		SessionID string
		Path      string
	}

	var rows []sessionEvent

	err := s.db.
		Model(&models.Event{}).
		Select("session_id, path").
		Where("property_id = ? AND session_id != '' AND created_at BETWEEN ? AND ?", propertyID, start, end).
		Order("session_id, created_at ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	sessions := make(map[string][]string)
	for _, row := range rows {
		sessions[row.SessionID] = append(sessions[row.SessionID], row.Path)
	}

	counts := make(map[string]int)
	for _, paths := range sessions {
		if len(paths) > 1 {
			counts[strings.Join(paths, " → ")]++
		}
	}

	results := make([]models.TopJourneyPath, 0, len(counts))
	for seq, count := range counts {
		results = append(results, models.TopJourneyPath{Sequence: seq, Count: count})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}
