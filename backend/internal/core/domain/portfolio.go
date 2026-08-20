package domain

import "time"

type Lot struct {
	TradeID              int       `json:"tradeId"`
	DepotID               int       `json:"depotId"`
	StockID              int       `json:"stockId"`
	DateOfPurchase       time.Time `json:"dateOfPurchase"`
	Quantity             float64   `json:"quantity"`
	Remaining            float64   `json:"remaining"`
	TotalInCents         int       `json:"totalInCents"`
	RemainingCostInCents int       `json:"remainingCostInCents"`
}

type Position struct {
	DepotID                int     `json:"depotId"`
	StockID                int     `json:"stockId"`
	WKN                    string  `json:"wkn"`
	Ticker                 string  `json:"ticker"`
	Quantity               float64 `json:"quantity"`
	InvestedInCents        int     `json:"investedInCents"`
	AvgPriceInCents        int     `json:"avgPriceInCents"`
	CurrentPriceInCents    int     `json:"currentPriceInCents"`
	CurrentValueInCents    int     `json:"currentValueInCents"`
	UnrealizedGainInCents  int     `json:"unrealizedGainInCents"`
	Lots                   []Lot   `json:"lots"`
}

type SellAllocation struct {
	SellTradeID         int       `json:"sellTradeId"`
	BuyTradeID          int       `json:"buyTradeId"`
	StockID             int       `json:"stockId"`
	Quantity            float64   `json:"quantity"`
	CostBasisInCents    int       `json:"costBasisInCents"`
	ProceedsInCents     int       `json:"proceedsInCents"`
	RealizedGainInCents int       `json:"realizedGainInCents"`
	BuyDate             time.Time `json:"buyDate"`
	SellDate            time.Time `json:"sellDate"`
}

type Portfolio struct {
	DepotID                int        `json:"depotId"`
	Positions              []Position `json:"positions"`
	InvestedInCents        int        `json:"investedInCents"`
	RealizedGainInCents    int        `json:"realizedGainInCents"`
	CurrentValueInCents    int        `json:"currentValueInCents"`
	UnrealizedGainInCents  int        `json:"unrealizedGainInCents"`
}

type TradeDTO struct {
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
	CostBasisInCents    int       `json:"costBasisInCents"`
	ProceedsInCents     int       `json:"proceedsInCents"`
	RealizedGainInCents int       `json:"realizedGainInCents"`
	CanDelete           bool      `json:"canDelete"`
}
