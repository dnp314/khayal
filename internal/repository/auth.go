package repository

import (
	"context"
	"fmt"
	"time"

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
		`select id, name, email, password_hash
		from users
		where name=($1)`,
		username,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		return entities.User{}, fmt.Errorf("get user by name %q: %w", username, err)
	}
	return user, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	_, err := r.db.Exec(
		ctx,
		`insert into sessions (id, user_id, expires_at)
		values ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert session for user %d: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) GetUserBySession(ctx context.Context, token string) (entities.User, error) {
	var user entities.User
	err := r.db.QueryRow(
		ctx,
		`select u.id, u.name, u.email, u.password_hash
		from sessions s
		join users u on u.id = s.user_id
		where s.id = $1 and s.expires_at > current_timestamp`,
		token,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		return entities.User{}, fmt.Errorf("get user by session: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	if _, err := r.db.Exec(
		ctx,
		`delete from sessions where id = $1`,
		token,
	); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
