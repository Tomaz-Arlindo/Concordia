package services

import (
	"context"
	"errors"
	"testing"

	"desktop/internal/models"
)

func TestAuthServiceValidatesRegistrationBeforeDatabaseAccess(t *testing.T) {
	service := NewAuthService(nil)
	tests := []struct {
		name  string
		input models.RegisterInput
		want  error
	}{
		{name: "nome vazio", input: models.RegisterInput{Email: "teste@example.com", Password: "SenhaTeste09"}, want: ErrInvalidInput},
		{name: "email vazio", input: models.RegisterInput{Name: "TESTE_", Password: "SenhaTeste09"}, want: ErrInvalidInput},
		{name: "senha curta", input: models.RegisterInput{Name: "TESTE_", Email: "teste@example.com", Password: "curta"}, want: ErrInvalidInput},
		{name: "email inválido", input: models.RegisterInput{Name: "TESTE_", Email: "teste@", Password: "SenhaTeste09"}, want: ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.Register(context.Background(), tt.input); !errors.Is(err, tt.want) {
				t.Fatalf("esperado %v, obtido %v", tt.want, err)
			}
		})
	}
}

func TestValidEmailAcceptsAndRejectsMailboxFormats(t *testing.T) {
	for _, email := range []string{"abc", "a@b", "teste@", "@example.com", "nome com espaço@example.com"} {
		if ValidEmail(email) {
			t.Errorf("e-mail inválido aceito: %q", email)
		}
	}
	for _, email := range []string{"teste@example.com", "nome.sobrenome+tag@sub.example.org"} {
		if !ValidEmail(email) {
			t.Errorf("e-mail válido recusado: %q", email)
		}
	}
}
