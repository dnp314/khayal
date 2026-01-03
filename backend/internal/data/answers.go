package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Answer struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question"`
	Answer     string    `json:"answer"`
	CreatedAt  time.Time `json:"created"`
}

type AnswerModel struct {
	DB *sql.DB
}

func (m *AnswerModel) Get(id int64) (*Answer, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, question_id, answer 
		FROM answers
		WHERE id=$1`

	var answer Answer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&answer.ID,
		&answer.QuestionID,
		&answer.Answer,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}
	return &answer, nil

}

func (m *AnswerModel) GetAll(id int64) ([]*Answer, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT a.id, a.question_id, q.title, a.answer
	FROM answers a
	LEFT JOIN questions q	
	ON a.question_id = q.id
	WHERE q.id=$1
	`

	answers := []*Answer{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, ErrRecordNotFound
	}

	defer rows.Close()

	for rows.Next() {

		var answer Answer
		var question Question

		err := rows.Scan(
			&answer.ID,
			&answer.QuestionID,
			&question.Title,
			&answer.Answer,
		)

		if err != nil {
			return nil, err
		}

		answers = append(answers, &answer)

	}

	return answers, nil
}

func (m *AnswerModel) Insert(answer *Answer, question *Question) error {

	// TODO: Rewrite with Transactions
	// NOTE: prepared statements do not allow multiple commands
	query1 := `
	INSERT INTO answers(question_id, answer)
	VALUES ($2, $1)
	RETURNING id, created_at;`

	query2 := `
	UPDATE questions
	SET scheduled_at = $1
	WHERE id = $2;`

	args1 := []any{answer.Answer, answer.QuestionID}
	args2 := []any{question.ScheduledAt, answer.QuestionID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query1, args1...).Scan(&answer.ID, &answer.CreatedAt)
	if err != nil {
		return err
	}

	_, err = m.DB.ExecContext(ctx, query2, args2...)
	if err != nil {
		return err
	}

	return nil
}
