package repository

import (
	"context"

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

func (r *AuthRepository) PostRegister(ctx context.Context, username string, password string) error {
	_, err := r.db.Exec(
		ctx,
		`insert into users (name ,passwordhash)
		values ($1, $2)`,
		username, password,
	)
	return err
}
