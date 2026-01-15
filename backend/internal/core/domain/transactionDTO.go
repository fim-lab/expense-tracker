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
