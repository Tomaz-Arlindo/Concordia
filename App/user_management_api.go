package main

import (
	"errors"
	"fmt"

	"desktop/internal/models"
	"desktop/internal/repositories"
	"desktop/internal/services"
)

func (a *App) requireAdmin() (*models.User, error) {
	user, err := a.authenticatedUser()
	if err != nil {
		return nil, err
	}
	if user.Role != models.RoleAdmin {
		return nil, errors.New("ação permitida somente para ADMIN")
	}
	return user, nil
}

func (a *App) userManagementService() *services.UserManagementService {
	return services.NewUserManagementService(repositories.NewUserRepository(a.db))
}

func roleDisplayName(role string) string {
	switch role {
	case models.RoleAdmin:
		return "Administrador"
	case models.RoleCurador:
		return "Curador"
	case models.RolePesquisador:
		return "Pesquisador"
	}
	return role
}

func (a *App) ListUsers() []models.User {
	if _, err := a.requireAdmin(); err != nil || a.db == nil {
		return []models.User{}
	}
	users, err := a.userManagementService().List(a.ctx)
	if err != nil {
		fmt.Printf("Concordia: falha ao listar usuários: %v\n", err)
		return []models.User{}
	}
	return users
}

func (a *App) UpdateUserRole(id uint64, role string) ActionResponse {
	if _, err := a.requireAdmin(); err != nil {
		return ActionResponse{Success: false, Message: "Somente o administrador pode alterar funções."}
	}
	if a.db == nil {
		return ActionResponse{Success: false, Message: "Banco de dados não conectado."}
	}

	user, err := a.userManagementService().ChangeRole(a.ctx, id, role)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidRole):
			return ActionResponse{Success: false, Message: "Função inválida. Use Pesquisador ou Curador."}
		case errors.Is(err, repositories.ErrUserNotFound):
			return ActionResponse{Success: false, Message: "Usuário não encontrado."}
		case errors.Is(err, services.ErrProtectedUser):
			return ActionResponse{Success: false, Message: "Administradores não podem ser alterados."}
		default:
			return ActionResponse{Success: false, Message: "Não foi possível alterar a função."}
		}
	}

	return ActionResponse{
		Success: true,
		Message: fmt.Sprintf("Função de %s alterada para %s.", user.Name, roleDisplayName(role)),
		ID:      id,
	}
}

func (a *App) DeleteUser(id uint64) ActionResponse {
	if _, err := a.requireAdmin(); err != nil {
		return ActionResponse{Success: false, Message: "Somente o administrador pode excluir usuários."}
	}
	if a.db == nil {
		return ActionResponse{Success: false, Message: "Banco de dados não conectado."}
	}

	user, err := a.userManagementService().Delete(a.ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrUserNotFound):
			return ActionResponse{Success: false, Message: "Usuário não encontrado."}
		case errors.Is(err, services.ErrProtectedUser):
			return ActionResponse{Success: false, Message: "Administradores não podem ser excluídos."}
		default:
			return ActionResponse{Success: false, Message: "Não foi possível excluir o usuário."}
		}
	}

	return ActionResponse{
		Success: true,
		Message: fmt.Sprintf("Usuário %s excluído.", user.Name),
		ID:      id,
	}
}