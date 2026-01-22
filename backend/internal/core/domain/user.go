package domain

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	SalaryCents  int    `json:"salaryCents"`
}
