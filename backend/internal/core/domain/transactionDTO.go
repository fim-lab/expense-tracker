package domain

import "time"

type TransactionDTO struct {
	ID            int             `json:"id"`
	Date          time.Time       `json:"date"`
	Description   string          `json:"description"`
	AmountInCents int             `json:"amountInCents"`
	Type          TransactionType `json:"type"`
	BudgetName    string          `json:"budgetName"`
	WalletName    string          `json:"walletName"`
	IsPending     bool            `json:"isPending"`
	IsDebt        bool            `json:"isDebt"`
}

type TransactionSearchCriteria struct {
	SearchTerm    *string
	FromDate      *time.Time
	UntilDate     *time.Time
	BudgetID      *int
	BudgetGroupID *int
	WalletID      *int
	Type          *TransactionType
	IsDebt        *bool
	Page          int
	PageSize      int
}

type PaginatedTransactions struct {
	Transactions []TransactionDTO `json:"transactions"`
	Total        int              `json:"total"`
	SumInCents   int              `json:"sumInCents"`
	Page         int              `json:"page"`
	PageSize     int              `json:"pageSize"`
}

type ImportTransaction struct {
	ID            string    `json:"id"`
	Date          time.Time `json:"date"`
	Budget        string    `json:"budget"`
	Wallet        string    `json:"wallet"`
	Description   string    `json:"description"`
	AmountInCents int       `json:"amountInCents"`
	Type          string    `json:"type"`
	IsPending     bool      `json:"isPending"`
	IsDebt        bool      `json:"isDebt"`
}

type ImportBudget struct {
	Name         string `json:"name"`
	ValueInCents int    `json:"valueInCents"`
	Account      string `json:"account"`
}

type ImportWallet struct {
	Name    string `json:"name"`
	IsDepot bool   `json:"isDepot"`
}

type ImportSettings struct {
	Gehalt  int            `json:"gehalt"`
	Budgets []ImportBudget `json:"budgets"`
	Wallets []ImportWallet `json:"wallets"`
}

type FullImportData struct {
	Settings     ImportSettings      `json:"settings"`
	Transactions []ImportTransaction `json:"transactions"`
}
