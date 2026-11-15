package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"desktop/internal/models"
	"desktop/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email já cadastrado")
	ErrInvalidCredentials = errors.New("email ou senha inválidos")
	ErrInvalidInput       = errors.New("dados inválidos")
)

type AuthService struct {
	users *repositories.UserRepository
}

var ErrInvalidEmail = errors.New("email fora do formato válido")

func NewAuthService(users *repositories.UserRepository) *AuthService {
	return &AuthService{users: users}
}

// ValidEmail accepts one mailbox with a dotted domain and no display name.
func ValidEmail(email string) bool {
	if strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || strings.Count(email, "@") != 1 {
		return false
	}
	local, domain, ok := strings.Cut(email, "@")
	return ok && local != "" && strings.Contains(domain, ".") && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}

func (s *AuthService) Register(ctx context.Context, input models.RegisterInput) (*models.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Name == "" || input.Email == "" || len(input.Password) < 8 {
		return nil, ErrInvalidInput
	}
	if !ValidEmail(input.Email) {
		return nil, ErrInvalidEmail
	}

	_, err := s.users.FindByEmail(ctx, input.Email)
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
		Role:         models.RolePesquisador,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
