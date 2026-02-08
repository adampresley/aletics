package models

type CountryLookup struct {
	Continent *Continent `json:"continent"`
	Country   *Country   `json:"country"`
}

type Continent struct {
	Code  *string           `json:"code"`
	Names map[string]string `json:"names"`
}

type Country struct {
	IsoCode *string           `json:"iso_code"`
	Names   map[string]string `json:"names"`
}
