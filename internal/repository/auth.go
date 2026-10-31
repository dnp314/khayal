package repository

import (
	"context"
	"fmt"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/jackc/pgx/v5"
)

type AuthRepository struct {
	db *pgx.Conn
}

func NewAuthRepostiory(db *pgx.Conn) *AuthRepository {
	return &AuthRepository{
		db: db,
	}

}

func (r *AuthRepository) PostRegister(ctx context.Context, username, password, email string) error {
	_, err := r.db.Exec(
		ctx,
		`insert into users (name ,password_hash, email)
		values ($1, $2, $3)`,
		username, password, email,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetUserByName(ctx context.Context, username string) (entities.User, error) {
	var user entities.User
	err := r.db.QueryRow(
		ctx,
		`select id, name, password_hash
		from users
		where name=($1)`,
		username,
	).Scan(
		&user.ID,
		&user.Name,
		&user.PasswordHash,
	)
	if err != nil {
		return entities.User{}, fmt.Errorf("get user by name %q: %w", username, err)
	}
	return user, nil
}
