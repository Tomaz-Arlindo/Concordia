package services

import (
	"context"
	"errors"
	"testing"

	"desktop/internal/models"
)

func TestAdminSetupValidatesInputBeforeDatabaseAccess(t *testing.T) {
	service := NewAdminSetupService(nil)
	tests := []struct {
		name  string
		input models.RegisterInput
		want  error
	}{
		{name: "nome vazio", input: models.RegisterInput{Email: "admin@example.com", Password: "SenhaTeste09"}, want: ErrInvalidInput},
		{name: "email vazio", input: models.RegisterInput{Name: "TESTE_Admin", Password: "SenhaTeste09"}, want: ErrInvalidInput},
		{name: "senha curta", input: models.RegisterInput{Name: "TESTE_Admin", Email: "admin@example.com", Password: "curta"}, want: ErrInvalidInput},
		{name: "email inválido", input: models.RegisterInput{Name: "TESTE_Admin", Email: "admin@", Password: "SenhaTeste09"}, want: ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.CreateInitialAdmin(context.Background(), tt.input); !errors.Is(err, tt.want) {
				t.Fatalf("esperado %v, obtido %v", tt.want, err)
			}
		})
	}
}
