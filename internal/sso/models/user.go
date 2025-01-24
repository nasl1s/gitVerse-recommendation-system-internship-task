package models

import "time"

type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john.doe@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"john.doe@example.com"`
	Password string `json:"password" example:"securepassword123"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type UserSSO struct {
	ID           int64     `json:"id" example:"1"`
	Name         string    `json:"name" example:"John Doe"`
	Email        string    `json:"email" example:"john.doe@example.com"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}
