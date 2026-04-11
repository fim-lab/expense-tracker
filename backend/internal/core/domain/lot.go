package domain

import "time"

type Lot struct {
	ID             int       `json:"id"`
	DateOfPurchase time.Time `json:"dateOfPurchase"`
	WKN            string    `json:"wkn"`
	Amount         float64   `json:"amount"`
	Remaining      float64   `json:"remaining"`
	DepotID        int       `json:"depotId"`
	PriceInCents   int       `json:"priceInCents"`
}
