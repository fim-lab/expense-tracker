package domain

import "github.com/google/uuid"

type Wallet struct {
	ID      uuid.UUID `json:"id"`
	UserID  int       `json:"userId"`
	Name    string    `json:"name"`
	Balance int64     `json:"balance"`
}
