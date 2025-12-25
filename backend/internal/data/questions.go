package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Question struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description sql.NullString `json:"description,omitempty"`
	IsAnswered  bool           `json:"-"`
	ScheduledAt time.Time      `json:"scheduledAt"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type QuestionModel struct {
	DB *sql.DB
}

func (m *QuestionModel) Get(id int64) (*Question, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT  id, created_at, title, description
		FROM questions
		WHERE id = $1`

	var question Question

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&question.ID,
		&question.CreatedAt,
		&question.Title,
		&question.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	return &question, nil

}

func (m *QuestionModel) GetAll(title string, genres []string, filters Filters) ([]*Question, MetaData, error) {

	// WARNING: this does not seem to work, with or without the title
	query :=
		`SELECT count(*) OVER(),id, created_at, title
		FROM questions`

	// if the query takes longer than 3 seconds, it would cancel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases the resources

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, MetaData{}, err
	}

	// what does this do?
	// closes the connection to the database
	// only used with query or query context, with queryrow it is auto managed
	defer rows.Close()

	totalRecords := 0
	questions := []*Question{}

	for rows.Next() {

		var question Question

		err := rows.Scan(
			&totalRecords,
			&question.ID,
			&question.CreatedAt,
			&question.Title,
		)

		if err != nil {
			return nil, MetaData{}, err
		}

		questions = append(questions, &question)
	}

	if err := rows.Err(); err != nil {
		return nil, MetaData{}, err
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.Pagesize)

	return questions, metadata, nil

}

func (m *QuestionModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (m *QuestionModel) Insert(question *Question) error {

	query := `
	INSERT INTO questions(title, description)
	VALUES($1, $2)
	RETURNING id, created_at
	`

	args := []interface{}{question.Title, question.Description}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&question.ID, &question.CreatedAt)

}
