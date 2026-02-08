package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/rester/clientoptions"
)

func TestGetCountryInfo(t *testing.T) {
	// Mock Maxmind GeoLite2 country response
	mockResponse := models.CountryLookup{
		Continent: &models.Continent{
			Code:  ptrStr("NA"),
			Names: map[string]string{"en": "North America"},
		},
		Country: &models.Country{
			IsoCode: ptrStr("US"),
			Names:   map[string]string{"en": "United States"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path matches expected format
		if r.URL.Path != "/country/8.8.8.8" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Verify basic auth is present
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth to be set")
		}
		if user != "test-account" || pass != "test-key" {
			t.Errorf("unexpected basic auth: %s / %s", user, pass)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	svc := NewIpLookupService(IpLookupServiceConfig{
		ApiAccountId: "test-account",
		ApiKey:       "test-key",
		RestConfig: clientoptions.New(
			server.URL,
			clientoptions.WithBasicAuth("test-account", "test-key"),
		),
	})

	result, err := svc.GetCountryInfo("8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify country
	if result.Country == nil {
		t.Fatal("expected country to be set")
	}
	if result.Country.IsoCode == nil || *result.Country.IsoCode != "US" {
		t.Errorf("expected country ISO code 'US', got %v", result.Country.IsoCode)
	}
	if name, ok := result.Country.Names["en"]; !ok || name != "United States" {
		t.Errorf("expected country name 'United States', got %v", result.Country.Names)
	}

	// Verify continent
	if result.Continent == nil {
		t.Fatal("expected continent to be set")
	}
	if result.Continent.Code == nil || *result.Continent.Code != "NA" {
		t.Errorf("expected continent code 'NA', got %v", result.Continent.Code)
	}
	if name, ok := result.Continent.Names["en"]; !ok || name != "North America" {
		t.Errorf("expected continent name 'North America', got %v", result.Continent.Names)
	}
}

func TestGetCountryInfo_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewIpLookupService(IpLookupServiceConfig{
		ApiAccountId: "test-account",
		ApiKey:       "test-key",
		RestConfig: clientoptions.New(
			server.URL,
			clientoptions.WithBasicAuth("test-account", "test-key"),
		),
	})

	_, err := svc.GetCountryInfo("8.8.8.8")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestGetCountryInfo_NilFields(t *testing.T) {
	// Response with no country or continent data
	mockResponse := models.CountryLookup{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	svc := NewIpLookupService(IpLookupServiceConfig{
		ApiAccountId: "test-account",
		ApiKey:       "test-key",
		RestConfig: clientoptions.New(
			server.URL,
			clientoptions.WithBasicAuth("test-account", "test-key"),
		),
	})

	result, err := svc.GetCountryInfo("192.0.2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Country != nil {
		t.Error("expected nil country for empty response")
	}
	if result.Continent != nil {
		t.Error("expected nil continent for empty response")
	}
}

func ptrStr(s string) *string {
	return &s
}
