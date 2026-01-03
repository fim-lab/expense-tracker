package domain

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrMissingDescription  = errors.New("description is required")
	ErrMissingBudget       = errors.New("budget name is required")
	ErrBudgetNotFound      = errors.New("budget not found or unauthorized")
	ErrMissingWallet       = errors.New("wallet name is required")
	ErrWalletNotFound      = errors.New("wallet not found or unauthorized")
	ErrUnauthorized        = errors.New("user not authorized")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrSessionNotFound     = errors.New("session not found")
)
