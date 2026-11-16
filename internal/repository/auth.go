package repository

import (
	"context"
	"fmt"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/jackc/pgx/v5"
)

const userColumns = `
	id,
	name,
	email,
	coalesce(password_hash, ''),
	coalesce(auth_key_hash, ''),
	coalesce(kdf_algorithm, ''),
	coalesce(kdf_iterations, 0),
	kdf_salt,
	kdf_version,
	wrapped_data_key,
	wrapped_data_key_iv,
	recovery_wrapped_data_key,
	recovery_wrapped_data_key_iv,
	recovery_verifier,
	crypto_setup_at`

type AuthRepository struct {
	db *pgx.Conn
}

func NewAuthRepostiory(db *pgx.Conn) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func scanUser(row pgx.Row) (entities.User, error) {
	var user entities.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AuthKeyHash,
		&user.KDFAlgorithm,
		&user.KDFIterations,
		&user.KDFSalt,
		&user.KDFVersion,
		&user.WrappedDataKey,
		&user.WrappedDataKeyIV,
		&user.RecoveryWrappedDataKey,
		&user.RecoveryDataKeyIV,
		&user.RecoveryVerifier,
		&user.CryptoSetupAt,
	)
	return user, err
}

func (r *AuthRepository) CreateUser(ctx context.Context, user entities.User) (int, error) {
	var id int
	err := r.db.QueryRow(
		ctx,
		`insert into users (
			name, email, password_hash,
			auth_key_hash, kdf_algorithm, kdf_iterations, kdf_salt, kdf_version,
			wrapped_data_key, wrapped_data_key_iv,
			recovery_wrapped_data_key, recovery_wrapped_data_key_iv,
			recovery_verifier, crypto_setup_at
		) values (
			$1, $2, nullif($3, ''),
			$4, $5, $6, $7, $8,
			$9, $10,
			$11, $12,
			$13, $14
		) returning id`,
		user.Name, user.Email, user.PasswordHash,
		user.AuthKeyHash, user.KDFAlgorithm, user.KDFIterations, user.KDFSalt, user.KDFVersion,
		user.WrappedDataKey, user.WrappedDataKeyIV,
		user.RecoveryWrappedDataKey, user.RecoveryDataKeyIV,
		user.RecoveryVerifier, user.CryptoSetupAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	user, err := scanUser(r.db.QueryRow(
		ctx,
		`select `+userColumns+` from users where lower(email) = lower($1)`,
		email,
	))
	if err != nil {
		return entities.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) GetUserByName(ctx context.Context, username string) (entities.User, error) {
	user, err := scanUser(r.db.QueryRow(
		ctx,
		`select `+userColumns+` from users where name = $1`,
		username,
	))
	if err != nil {
		return entities.User{}, fmt.Errorf("get user by name: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id int) (entities.User, error) {
	user, err := scanUser(r.db.QueryRow(
		ctx,
		`select `+userColumns+` from users where id = $1`,
		id,
	))
	if err != nil {
		return entities.User{}, fmt.Errorf("get user %d: %w", id, err)
	}
	return user, nil
}

func (r *AuthRepository) UpdateAuthKeys(
	ctx context.Context,
	userID int,
	authKeyHash, kdfAlgorithm string,
	kdfIterations int,
	salt, wrappedDataKey, wrappedDataKeyIV []byte,
) error {
	_, err := r.db.Exec(
		ctx,
		`update users
		 set auth_key_hash = $2,
		     kdf_algorithm = $3,
		     kdf_iterations = $4,
		     kdf_salt = $5,
		     wrapped_data_key = $6,
		     wrapped_data_key_iv = $7,
		     crypto_setup_at = coalesce(crypto_setup_at, now())
		 where id = $1`,
		userID, authKeyHash, kdfAlgorithm, kdfIterations,
		salt, wrappedDataKey, wrappedDataKeyIV,
	)
	if err != nil {
		return fmt.Errorf("update auth keys for user %d: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) SetupLegacyCrypto(
	ctx context.Context,
	userID int,
	authKeyHash, kdfAlgorithm string,
	kdfIterations int,
	salt, wrappedDataKey, wrappedDataKeyIV []byte,
	recoveryWrappedDataKey, recoveryWrappedDataKeyIV, recoveryVerifier []byte,
) error {
	_, err := r.db.Exec(
		ctx,
		`update users
		 set auth_key_hash = $2,
		     kdf_algorithm = $3,
		     kdf_iterations = $4,
		     kdf_salt = $5,
		     wrapped_data_key = $6,
		     wrapped_data_key_iv = $7,
		     recovery_wrapped_data_key = $8,
		     recovery_wrapped_data_key_iv = $9,
		     recovery_verifier = $10,
		     crypto_setup_at = now()
		 where id = $1`,
		userID, authKeyHash, kdfAlgorithm, kdfIterations,
		salt, wrappedDataKey, wrappedDataKeyIV,
		recoveryWrappedDataKey, recoveryWrappedDataKeyIV, recoveryVerifier,
	)
	if err != nil {
		return fmt.Errorf("setup legacy crypto for user %d: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) RecoverUser(
	ctx context.Context,
	userID int,
	authKeyHash, kdfAlgorithm string,
	kdfIterations int,
	salt, wrappedDataKey, wrappedDataKeyIV []byte,
) error {
	return r.UpdateAuthKeys(ctx, userID, authKeyHash, kdfAlgorithm, kdfIterations, salt, wrappedDataKey, wrappedDataKeyIV)
}

func (r *AuthRepository) DeleteSessionsForUser(ctx context.Context, userID int) error {
	if _, err := r.db.Exec(ctx, `delete from sessions where user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete sessions for user %d: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session entities.Session) error {
	_, err := r.db.Exec(
		ctx,
		`insert into sessions (id, user_id, expires_at)
		values ($1, $2, $3)`,
		session.ID, session.UserID, session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert session for user %d: %w", session.UserID, err)
	}
	return nil
}

func (r *AuthRepository) GetSession(ctx context.Context, token string) (entities.Session, error) {
	var session entities.Session
	err := r.db.QueryRow(
		ctx,
		`select id, user_id, expires_at, created_at
		from sessions
		where id = $1`,
		token,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		return entities.Session{}, fmt.Errorf("get session: %w", err)
	}
	return session, nil
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
