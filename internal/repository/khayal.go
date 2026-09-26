package repository

import (
	"context"
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
		"INSERT INTO khayals (content,question_id) VALUES ($1, $2)",
		khayal, questionID,
	)

	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		ctx,
		`update questions q
		 set answered = true, return_at=($2)
		 where q.id=($1)
			
		`,
		questionID, returnAt,
	)
	if err != nil {
		return err
	}

	return err
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
		`SELECT k.id, k.content, k.created_at
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
		return nil, err
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
			return nil, err
		}

		khayals = append(khayals, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return khayals, nil
}
