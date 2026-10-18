package repository

import (
	"context"
	"log"

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

func (r *AuthRepository) PostRegister(ctx context.Context, username string, password string) error {
	_, err := r.db.Exec(
		ctx,
		`insert into users (name ,passwordhash)
		values ($1, $2)`,
		username, password,
	)
	return err
}

func (r *AuthRepository) GetUserByName(ctx context.Context, username string) (entities.User, error) {
	var user entities.User
	err := r.db.QueryRow(
		ctx,
		`select id, name, passwordhash
		from users
		where name=($1)`,
		username,
	).Scan(
		&user.ID,
		&user.Name,
		&user.PasswordHash,
	)
	if err != nil {
		log.Println(err)
	}
	return user, nil
}
