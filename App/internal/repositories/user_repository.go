package repositories

import (
	"context"
	"database/sql"
	"errors"

	"desktop/internal/models"
)

var ErrUserNotFound = errors.New("usuário não encontrado")

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.DB.QueryRowContext(
		ctx,
		`SELECT id, name, email, password_hash, role, created_at
		 FROM users
		 WHERE email = ?
		 LIMIT 1`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	result, err := r.DB.ExecContext(
		ctx,
		`INSERT INTO users (name, email, password_hash, role)
		 VALUES (?, ?, ?, ?)`,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(id)

	return r.DB.QueryRowContext(
		ctx,
		`SELECT created_at FROM users WHERE id = ?`,
		user.ID,
	).Scan(&user.CreatedAt)
}
