package services

import (
	"context"
	"errors"
	"strings"
	"sync"

	"desktop/internal/models"
	"desktop/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

var ErrAdminAlreadyExists = errors.New("já existe um administrador")

// mutex pra usar somente um goroutine
var adminSetupMu sync.Mutex

type AdminSetupService struct {
	users *repositories.UserRepository
}

func NewAdminSetupService(users *repositories.UserRepository) *AdminSetupService {
	return &AdminSetupService{users: users}
}

// devolve true se não existir admin
func (s *AdminSetupService) NeedsSetup(ctx context.Context) (bool, error) {
	var count int
	err := s.users.DB.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM users WHERE role = ?`,
		models.RoleAdmin,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *AdminSetupService) CreateInitialAdmin(ctx context.Context, input models.RegisterInput) (*models.User, error) {
	adminSetupMu.Lock()
	defer adminSetupMu.Unlock()

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Name == "" || input.Email == "" || len(input.Password) < 8 {
		return nil, ErrInvalidInput
	}

	needs, err := s.NeedsSetup(ctx)
	if err != nil {
		return nil, err
	}
	if !needs {
		return nil, ErrAdminAlreadyExists
	}

	_, err = s.users.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}
	if !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         models.RoleAdmin,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}