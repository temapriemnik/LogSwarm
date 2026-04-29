package domain

import "time"

// User represents a user in the system.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	IsActive     bool
}