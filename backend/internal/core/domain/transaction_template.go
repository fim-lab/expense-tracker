package domain

import "fmt"

type TransactionTemplate struct {
	ID            int             `json:"id"`
	UserID        int             `json:"userId"`
	Day           int             `json:"day"` // Day of the month (1-31)
	BudgetID      *int            `json:"budgetId"`
	WalletID      int             `json:"walletId"`
	Description   string          `json:"description"`
	AmountInCents int             `json:"amountInCents"`
	Type          TransactionType `json:"type"`
	Tags          []string        `json:"tags,omitempty"`
}

func (tt *TransactionTemplate) Validate() error {
	if tt.UserID == 0 {
		return fmt.Errorf("transaction template must have a user ID")
	}
	if tt.Day < 1 || tt.Day > 31 {
		return fmt.Errorf("day must be between 1 and 31")
	}
	if tt.WalletID == 0 {
		return fmt.Errorf("transaction template must have a wallet ID")
	}
	if tt.Description == "" {
		return fmt.Errorf("transaction template must have a description")
	}
	if tt.AmountInCents <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if tt.Type != Income && tt.Type != Expense {
		return fmt.Errorf("invalid transaction type")
	}
	return nil
}
