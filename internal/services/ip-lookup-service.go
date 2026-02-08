package services

import (
	"fmt"

	"github.com/adampresley/aletics/internal/models"
	"github.com/adampresley/rester"
	"github.com/adampresley/rester/clientoptions"
)

const (
	MaxmindBaseUrl string = "https://geolite.info/geoip/v2.1"
)

type IpLookupService struct {
	apiAccountId string
	apiKey       string
	restConfig   *clientoptions.ClientOptions
}

type IpLookupServiceConfig struct {
	ApiAccountId string
	ApiKey       string
	RestConfig   *clientoptions.ClientOptions
}

func NewIpLookupService(config IpLookupServiceConfig) *IpLookupService {
	return &IpLookupService{
		apiAccountId: config.ApiAccountId,
		apiKey:       config.ApiKey,
		restConfig:   config.RestConfig,
	}
}

func (s *IpLookupService) GetCountryInfo(ip string) (*models.CountryLookup, error) {
	var (
		err    error
		result = &models.CountryLookup{}
	)

	result, _, err = rester.Get[*models.CountryLookup](
		s.restConfig, fmt.Sprintf("/country/%s", ip),
	)

	if err != nil {
		return result, fmt.Errorf("error fetching country information for IP %s: %w", ip, err)
	}

	return result, nil
}
