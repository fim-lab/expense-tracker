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
}

type TransactionSearchCriteria struct {
	SearchTerm *string
	FromDate   *time.Time
	UntilDate  *time.Time
	BudgetID   *int
	WalletID   *int
	Type       *TransactionType
	Page       int
	PageSize   int
}

type PaginatedTransactions struct {
	Transactions []TransactionDTO `json:"transactions"`
	Total        int              `json:"total"`
	Page         int              `json:"page"`
	PageSize     int              `json:"pageSize"`
}
