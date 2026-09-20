package domain

import "time"

type Stock struct {
	ID           int       `json:"id"`
	WKN          string    `json:"wkn"`
	Ticker       string    `json:"ticker"`
	PriceInCents int       `json:"priceInCents"`
	LastFetched  time.Time `json:"lastFetched"`
}

type StockPriceRefresh struct {
	NeedsConfirmation bool  `json:"needsConfirmation"`
	Stock             Stock `json:"stock"`
	OldPriceInCents   int   `json:"oldPriceInCents"`
	NewPriceInCents   int   `json:"newPriceInCents"`
}
