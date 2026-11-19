package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/jackc/pgx/v5"
)

// why do i have methods over functions ?
type KhayalRepository struct {
	db *pgx.Conn
}

// why pointer, instead of copy
func NewKhayalRepostiory(db *pgx.Conn) *KhayalRepository {
	return &KhayalRepository{
		db: db,
	}
}

func (r *KhayalRepository) CreateKhayal(ctx context.Context, questionID int, khayal string, returnAt time.Time) error {
	_, err := r.db.Exec(
		ctx,
		`WITH inserted_khayal AS (
			INSERT INTO khayals (khayal, question_id)
			VALUES ($2, $1)
			RETURNING id
		)
		UPDATE questions
		SET answered = true, return_at = $3
		WHERE id = $1`,
		questionID, khayal, returnAt,
	)
	if err != nil {
		return fmt.Errorf("insert khayal and update question %d return time: %w", questionID, err)
	}

	return nil
}

// func (r *KhayalRepository) GetKhayal(ctx context.Context, id int) (entities.Khayal, error) {
// 	var khayal entities.Khayal
// 	err := r.db.QueryRow(
// 		ctx,
// 		"SELECT khayal,answered FROM khayals where id=($1)",
// 		id,
// 	).Scan(
// 		&khayal.Khayal,
// 		&khayal.Answered,
// 	)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	return khayal, err
// }

func (r *KhayalRepository) GetAllKhayals(ctx context.Context, questionID int) ([]entities.Khayal, error) {

	rows, err := r.db.Query(ctx,
		`SELECT k.id, k.khayal, k.created_at
			FROM khayals k
			JOIN questions q ON q.id = k.question_id
			WHERE k.question_id = $1
			  AND q.return_at > CURRENT_TIMESTAMP
			  AND (
			      SELECT MAX(k2.created_at)
			      FROM khayals k2
			      WHERE k2.question_id = $1
			  ) > CURRENT_TIMESTAMP - INTERVAL '1 day'
			ORDER BY k.created_at DESC;
		`,
		questionID)
	if err != nil {
		return nil, fmt.Errorf("query khayals for question %d: %w", questionID, err)
	}
	defer rows.Close()

	var khayals []entities.Khayal

	for rows.Next() {
		var q entities.Khayal

		err := rows.Scan(
			&q.ID,
			&q.Khayal,
			&q.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan khayal for question %d: %w", questionID, err)
		}

		khayals = append(khayals, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate khayals for question %d: %w", questionID, err)
	}

	return khayals, nil
}
