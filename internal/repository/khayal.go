package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type KhayalRepository struct {
	db *pgx.Conn
}

func NewKhayalRepostiory(db *pgx.Conn) *KhayalRepository {
	return &KhayalRepository{db: db}
}

func (r *KhayalRepository) CreateKhayal(
	ctx context.Context,
	userID, questionID int,
	clientID uuid.UUID,
	ciphertext []byte,
	encryptionVersion int,
	returnAt time.Time,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create khayal for question %d: %w", questionID, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(
		ctx,
		`insert into khayals (question_id, client_id, khayal_ciphertext, encryption_version)
		 values ($1, $2, $3, $4)`,
		questionID, clientID, ciphertext, encryptionVersion,
	)
	if err != nil {
		return fmt.Errorf("insert khayal for question %d: %w", questionID, err)
	}

	tag, err := tx.Exec(
		ctx,
		`update questions
		 set answered = true, return_at = $3
		 where id = $1 and user_id = $2`,
		questionID, userID, returnAt,
	)
	if err != nil {
		return fmt.Errorf("mark question %d answered: %w", questionID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("question %d does not belong to user %d", questionID, userID)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create khayal for question %d: %w", questionID, err)
	}
	return nil
}

func (r *KhayalRepository) GetAllKhayals(ctx context.Context, userID, questionID int) ([]entities.Khayal, error) {
	rows, err := r.db.Query(ctx,
		`select k.id, q.user_id, k.question_id, k.client_id, k.khayal_ciphertext,
		        k.encryption_version, k.created_at
		 from khayals k
		 join questions q on q.id = k.question_id
		 where k.question_id = $1
		   and q.user_id = $2
		   and k.khayal_ciphertext is not null
		   and q.return_at > current_timestamp
		   and (
		       select max(k2.created_at)
		       from khayals k2
		       where k2.question_id = $1
		   ) > current_timestamp - interval '1 day'
		 order by k.created_at desc`,
		questionID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query khayals for question %d: %w", questionID, err)
	}
	defer rows.Close()

	var khayals []entities.Khayal
	for rows.Next() {
		var k entities.Khayal
		if err := rows.Scan(
			&k.ID,
			&k.UserID,
			&k.QuestionID,
			&k.ClientID,
			&k.KhayalCiphertext,
			&k.EncryptionVersion,
			&k.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan khayal for question %d: %w", questionID, err)
		}
		khayals = append(khayals, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate khayals for question %d: %w", questionID, err)
	}
	return khayals, nil
}

func (r *KhayalRepository) GetLegacyKhayals(ctx context.Context, userID int) ([]entities.LegacyKhayal, error) {
	rows, err := r.db.Query(
		ctx,
		`select k.id, k.question_id, k.khayal
		 from khayals k
		 join questions q on q.id = k.question_id
		 where q.user_id = $1 and k.khayal_ciphertext is null and k.khayal is not null
		 order by k.id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query legacy khayals for user %d: %w", userID, err)
	}
	defer rows.Close()

	var khayals []entities.LegacyKhayal
	for rows.Next() {
		var k entities.LegacyKhayal
		if err := rows.Scan(&k.ID, &k.QuestionID, &k.Khayal); err != nil {
			return nil, fmt.Errorf("scan legacy khayal for user %d: %w", userID, err)
		}
		khayals = append(khayals, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy khayals for user %d: %w", userID, err)
	}
	return khayals, nil
}

func (r *KhayalRepository) ImportLegacyKhayal(
	ctx context.Context,
	userID, id, questionID int,
	clientID uuid.UUID,
	ciphertext []byte,
	encryptionVersion int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`update khayals k
		 set client_id = $4,
		     khayal_ciphertext = $5,
		     encryption_version = $6,
		     khayal = null
		 from questions q
		 where k.id = $1
		   and k.question_id = $2
		   and q.id = k.question_id
		   and q.user_id = $3
		   and k.khayal_ciphertext is null`,
		id, questionID, userID, clientID, ciphertext, encryptionVersion,
	)
	if err != nil {
		return fmt.Errorf("import legacy khayal %d for user %d: %w", id, userID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("legacy khayal %d for user %d was not found or already migrated", id, userID)
	}
	return nil
}
