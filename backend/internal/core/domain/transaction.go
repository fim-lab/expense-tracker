package domain

import "time"

type TransactionType string

const (
	Income  TransactionType = "INCOME"
	Expense TransactionType = "EXPENSE"
)

type Transaction struct {
	ID            int             `json:"id"`
	UserID        int             `json:"userId"`
	Date          time.Time       `json:"date"`
	BudgetID      int             `json:"budgetId"`
	WalletID      int             `json:"walletId"`
	Description   string          `json:"description"`
	AmountInCents int             `json:"amountInCents"`
	Type          TransactionType `json:"type"`
	IsPending     bool            `json:"isPending"`
	IsDebt        *bool           `json:"isDebt,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
}
