package services

import (
	"context"
	"errors"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

var (
	ErrInvalidRole   = errors.New("função inválida")
	ErrProtectedUser = errors.New("administradores não podem ser alterados ou excluídos")
)

type UserManagementService struct {
	users *repositories.UserRepository
}

func NewUserManagementService(users *repositories.UserRepository) *UserManagementService {
	return &UserManagementService{users: users}
}

func (s *UserManagementService) List(ctx context.Context) ([]models.User, error) {
	return s.users.ListAll(ctx)
}

// só aceita pesquisador ou curador nunca admin
func (s *UserManagementService) ChangeRole(ctx context.Context, id uint64, role string) (*models.User, error) {
	if role != models.RolePesquisador && role != models.RoleCurador {
		return nil, ErrInvalidRole
	}

	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Role == models.RoleAdmin {
		return nil, ErrProtectedUser
	}
	if user.Role == role {
		return user, nil
	}

	changed, err := s.users.UpdateRole(ctx, id, role)
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, ErrProtectedUser
	}

	user.Role = role
	return user, nil
}

func (s *UserManagementService) Delete(ctx context.Context, id uint64) (*models.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Role == models.RoleAdmin {
		return nil, ErrProtectedUser
	}

	deleted, err := s.users.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, ErrProtectedUser
	}
	return user, nil
}