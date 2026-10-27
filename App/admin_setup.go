package main

import (
	"errors"
	"fmt"

	"desktop/internal/models"
	"desktop/internal/repositories"
	"desktop/internal/services"
)

func (a *App) adminSetupService() *services.AdminSetupService {
	return services.NewAdminSetupService(repositories.NewUserRepository(a.db))
}

func (a *App) NeedsAdminSetup() bool {
	if a.db == nil {
		return false
	}
	needs, err := a.adminSetupService().NeedsSetup(a.ctx)
	if err != nil {
		fmt.Printf("Concordia: falha ao verificar administrador: %v\n", err)
		return false
	}
	return needs
}

func (a *App) CreateInitialAdmin(name, email, password string) AuthResponse {
	if a.db == nil {
		return AuthResponse{Success: false, Message: "Banco de dados não conectado."}
	}

	user, err := a.adminSetupService().CreateInitialAdmin(a.ctx, models.RegisterInput{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return AuthResponse{Success: false, Message: "Preencha nome e email e use uma senha com pelo menos 8 caracteres."}
		case errors.Is(err, services.ErrEmailAlreadyExists):
			return AuthResponse{Success: false, Message: "Este email já está cadastrado. Use outro email."}
		case errors.Is(err, services.ErrAdminAlreadyExists):
			return AuthResponse{Success: false, Message: "Já existe um administrador cadastrado."}
		default:
			return AuthResponse{Success: false, Message: "Não foi possível criar o administrador."}
		}
	}

	a.setCurrentUser(user)
	return AuthResponse{Success: true, Message: "Administrador criado com sucesso.", User: user}
}