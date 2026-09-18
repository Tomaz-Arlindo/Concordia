package models

import "time"

const (
	RolePesquisador = "PESQUISADOR"
	RoleCurador     = "CURADOR"
)

type User struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}
