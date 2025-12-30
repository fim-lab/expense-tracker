package domain

import "time"

type TransactionType string

const (
	Income  TransactionType = "INCOME"
	Expense TransactionType = "EXPENSE"
)

type Transaction struct {
	ID            string          `json:"id"`
	Date          time.Time       `json:"date"`
	Budget        string          `json:"budget"`
	Description   string          `json:"description"`
	AmountInCents int64           `json:"amountInCents"`
	Wallet        string          `json:"wallet"`
	Type          TransactionType `json:"type"`
	IsPending     bool            `json:"isPending"`
	IsDebt        *bool           `json:"isDebt,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
}