package domain

import "github.com/google/uuid"

type Budget struct {
	ID         uuid.UUID `json:"id"`
	UserID     int       `json:"userId"`
	Name       string    `json:"name"`
	LimitCents int64     `json:"limitCents"`
}
