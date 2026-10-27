package repositories

import (
	"context"
	"database/sql"
	"errors"

	"desktop/internal/models"
)

func (r *UserRepository) ListAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.DB.QueryContext(
		ctx,
		`SELECT id, name, email, role, created_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	var user models.User

	err := r.DB.QueryRowContext(
		ctx,
		`SELECT id, name, email, role, created_at FROM users WHERE id = ? LIMIT 1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// nunca altera admin e retorna false se nada foi alterado
func (r *UserRepository) UpdateRole(ctx context.Context, id uint64, role string) (bool, error) {
	result, err := r.DB.ExecContext(
		ctx,
		`UPDATE users SET role = ? WHERE id = ? AND role <> ?`,
		role, id, models.RoleAdmin,
	)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

// nunca remove admin
func (r *UserRepository) Delete(ctx context.Context, id uint64) (bool, error) {
	result, err := r.DB.ExecContext(
		ctx,
		`DELETE FROM users WHERE id = ? AND role <> ?`,
		id, models.RoleAdmin,
	)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}