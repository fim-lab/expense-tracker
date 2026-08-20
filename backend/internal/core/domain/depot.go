package domain

type Depot struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	UserID   int    `json:"userId"`
	WalletID int    `json:"walletId"`
	BudgetID int    `json:"budgetId"`
}

type DepotDTO struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	WalletID        int    `json:"walletId"`
	BudgetID        int    `json:"budgetId"`
	InvestedInCents int    `json:"investedInCents"`
}
