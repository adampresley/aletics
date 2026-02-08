package models

/*
This file contains data structures used for reporting and analytics.
These are not database models, but rather structures to hold the
results of aggregate queries.
*/

// ViewsOverTimeItem represents a single data point for a views-over-time graph.
type ViewsOverTimeItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// TopPathItem holds the count of views for a specific path.
type TopPathItem struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// BrowserCountItem holds the count of views for a specific browser.
type BrowserCountItem struct {
	Browser string `json:"browser"`
	Count   int    `json:"count"`
}
