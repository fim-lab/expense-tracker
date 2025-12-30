package ports

import "github.com/fim-lab/expense-tracker/backend/internal/core/domain"

type ExpenseService interface {
	CreateTransaction(t domain.Transaction) error
	GetTransactions() ([]domain.Transaction, error)
	DeleteTransaction(id string) error
}

type ExpenseRepository interface {
	Save(t domain.Transaction) error
	FindAll() ([]domain.Transaction, error)
	Delete(id string) error
}