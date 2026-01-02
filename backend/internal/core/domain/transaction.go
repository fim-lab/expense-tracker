package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	Income  TransactionType = "INCOME"
	Expense TransactionType = "EXPENSE"
)

type Transaction struct {
	ID            uuid.UUID       `json:"id"`
	UserID        int             `json:"userId"`
	Date          time.Time       `json:"date"`
	BudgetID      uuid.UUID       `json:"budgetId"`
	Description   string          `json:"description"`
	AmountInCents int64           `json:"amountInCents"`
	Wallet        string          `json:"wallet"`
	Type          TransactionType `json:"type"`
	IsPending     bool            `json:"isPending"`
	IsDebt        *bool           `json:"isDebt,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
}
