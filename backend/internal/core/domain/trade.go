package domain

import "time"

type TradeType string

const (
	TradeTypeBuy  TradeType = "BUY"
	TradeTypeSell TradeType = "SELL"
)

type Trade struct {
	ID                  int       `json:"id"`
	DepotID             int       `json:"depotId"`
	WalletTransactionID *int      `json:"walletTransactionId"`
	WKN                 string    `json:"wkn"`
	Type                TradeType `json:"type"`
	Quantity            float64   `json:"quantity"`
	PriceInCents        int       `json:"priceInCents"`
	FeesInCents         int       `json:"feesInCents"`
	TaxesInCents        int       `json:"taxesInCents"`
	Timestamp           time.Time `json:"timestamp"`
}
