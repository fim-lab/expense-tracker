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
	StockID             int       `json:"stockId"`
	WKN                 string    `json:"wkn"`
	Type                TradeType `json:"type"`
	Quantity            float64   `json:"quantity"`
	TotalInCents        int       `json:"totalInCents"`
	FeesInCents         int       `json:"feesInCents"`
	TaxesInCents        int       `json:"taxesInCents"`
	Timestamp           time.Time `json:"timestamp"`
}

// CashFlowInCents is the amount that moves on the depot's wallet for this trade.
// Fees and taxes are stored but deliberately not part of the cash flow yet;
// this is the single place to add them once that feature is wanted.
func (t Trade) CashFlowInCents() int {
	return t.TotalInCents
}
