package repository

import (
	"context"
	"log"

	"github.com/dnp314/khayal/internal/entities"

	"github.com/jackc/pgx/v5"
)

// why do i have methods over functions ?
type QuestionRepository struct {
	db *pgx.Conn
}

// why pointer, instead of copy
func NewQuestionRepostiory(db *pgx.Conn) *QuestionRepository {
	return &QuestionRepository{
		db: db,
	}

}

func (r *QuestionRepository) CreateQuestion(ctx context.Context, question string) error {
	_, err := r.db.Exec(
		ctx,
		"INSERT INTO questions (question) VALUES ($1)",
		question,
	)

	return err
}

func (r *QuestionRepository) GetQuestion(ctx context.Context, id int) (entities.Question, error) {
	var question entities.Question
	err := r.db.QueryRow(
		ctx,
		"SELECT id,question,answered FROM questions where id=($1)",
		id,
	).Scan(
		&question.ID,
		&question.Question,
		&question.Answered,
	)
	if err != nil {
		log.Println(err)
	}

	return question, err
}

func (r *QuestionRepository) GetAllQuestions(ctx context.Context) ([]entities.Question, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT id, question, answered, created_at, updated_at
		 FROM questions
		 where return_at is null or  return_at <= current_timestamp
		 order by return_at, created_at desc
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []entities.Question

	for rows.Next() {
		var q entities.Question

		err := rows.Scan(
			&q.ID,
			&q.Question,
			&q.Answered,
			&q.CreatedAt,
			&q.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
