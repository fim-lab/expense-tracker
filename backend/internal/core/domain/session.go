package domain

import "time"

type Session struct {
	UserID       int       `json:"user_id"`
	SessionToken string    `json:"token"`
	Expiry       time.Time `json:"expiry"`
}
