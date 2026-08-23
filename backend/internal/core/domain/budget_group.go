package domain

type BudgetGroup struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	Name   string `json:"name"`
}
