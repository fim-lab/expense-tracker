package services

import (
	"errors"
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/backend/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
)

func TestService_CreateTransaction(t *testing.T) {
	repo := memory.NewRepository()
	service := NewExpenseService(repo)

	tests := []struct {
		name    string
		tx      domain.Transaction
		wantErr error
	}{
		{
			name: "valid transaction",
			tx: domain.Transaction{
				ID:            "valid-1",
				Description:   "Groceries",
				AmountInCents: 5000,
				Budget:        "Food",
				Date:          time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid amount",
			tx: domain.Transaction{
				ID:            "invalid-1",
				Description:   "Free stuff?",
				AmountInCents: 0,
				Budget:        "Food",
			},
			wantErr: domain.ErrInvalidAmount,
		},
		{
			name: "missing description",
			tx: domain.Transaction{
				ID:            "invalid-2",
				Description:   "",
				AmountInCents: 100,
				Budget:        "Food",
			},
			wantErr: domain.ErrMissingDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateTransaction(tt.tx)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestService_GetTransactions(t *testing.T) {
	repo := memory.NewRepository()
	service := NewExpenseService(repo)

	_ = service.CreateTransaction(domain.Transaction{
		ID: "1", Description: "A", AmountInCents: 10, Budget: "B",
	})
	_ = service.CreateTransaction(domain.Transaction{
		ID: "2", Description: "B", AmountInCents: 20, Budget: "B",
	})

	txs, _ := service.GetTransactions()
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}
}
