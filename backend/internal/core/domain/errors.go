package domain

import "errors"

var (
	ErrInvalidAmount      = errors.New("amount must be greater than zero")
	ErrMissingDescription = errors.New("description is required")
	ErrMissingBudget      = errors.New("budget name is required")
	ErrTransactionNotFound = errors.New("transaction not found")
)