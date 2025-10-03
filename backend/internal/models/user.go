package models

import "time"

// User represents an account that can authenticate with the API.
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Sanitized returns a copy of the user safe for serialization.
func (u User) Sanitized() User {
	u.PasswordHash = ""
	return u
}
