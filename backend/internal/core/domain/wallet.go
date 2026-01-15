package domain

type Wallet struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	Name         string `json:"name"`
	BalanceCents int    `json:"balanceCents"`
}
