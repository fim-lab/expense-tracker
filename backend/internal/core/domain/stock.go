package domain

import "time"

type Stock struct {
	ID             int       `json:"id"`
	DateOfPurchase time.Time `json:"dateOfPurchase"`
	WKN            string    `json:"wkn"`
	Amount         float64   `json:"amount"`
	DepotID        int       `json:"depotId"`
	UserID         int       `json:"userId"`
	PriceInCents   int       `json:"priceInCents"`
}
