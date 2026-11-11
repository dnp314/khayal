package repository

import (
	"context"
	"fmt"

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

func (r *QuestionRepository) CreateQuestion(ctx context.Context, userID int, question string) error {
	_, err := r.db.Exec(
		ctx,
		"INSERT INTO questions (user_id, question) VALUES ($1, $2)",
		userID, question,
	)
	if err != nil {
		return fmt.Errorf("insert question for user %d: %w", userID, err)
	}

	return nil
}

func (r *QuestionRepository) GetQuestion(ctx context.Context, userID, id int) (entities.Question, error) {
	var question entities.Question
	err := r.db.QueryRow(
		ctx,
		"SELECT id, user_id, question, answered FROM questions where id=($1) and user_id=($2)",
		id, userID,
	).Scan(
		&question.ID,
		&question.UserID,
		&question.Question,
		&question.Answered,
	)
	if err != nil {
		return entities.Question{}, fmt.Errorf("get question %d for user %d: %w", id, userID, err)
	}

	return question, nil
}

func (r *QuestionRepository) GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT id, user_id, question, answered, created_at, updated_at
		 FROM questions
		 where user_id = $1
		   and (return_at is null or return_at <= current_timestamp)
		 order by return_at, created_at desc
		`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query questions for user %d: %w", userID, err)
	}
	defer rows.Close()

	var questions []entities.Question

	for rows.Next() {
		var q entities.Question

		err := rows.Scan(
			&q.ID,
			&q.UserID,
			&q.Question,
			&q.Answered,
			&q.CreatedAt,
			&q.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan question for user %d: %w", userID, err)
		}

		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questions for user %d: %w", userID, err)
	}

	return questions, nil
}
