package domain

type Budget struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	Name         string `json:"name"`
	LimitCents   int    `json:"limitCents"`
	BalanceCents int    `json:"balanceCents"`
}
