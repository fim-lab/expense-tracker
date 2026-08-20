package domain

import "time"

type Stock struct {
	ID           int       `json:"id"`
	WKN          string    `json:"wkn"`
	Ticker       string    `json:"ticker"`
	PriceInCents int       `json:"priceInCents"`
	LastFetched  time.Time `json:"lastFetched"`
}
