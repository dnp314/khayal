package repository

import (
	"context"
	"fmt"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QuestionRepository struct {
	db *pgx.Conn
}

func NewQuestionRepostiory(db *pgx.Conn) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) CreateQuestion(ctx context.Context, userID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error {
	_, err := r.db.Exec(
		ctx,
		`insert into questions (user_id, client_id, question_ciphertext, encryption_version)
		 values ($1, $2, $3, $4)`,
		userID, clientID, ciphertext, encryptionVersion,
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
		`select id, user_id, client_id, question_ciphertext, encryption_version,
		        answered, coalesce(return_at, to_timestamp(0)), created_at, updated_at
		 from questions
		 where id = $1 and user_id = $2 and question_ciphertext is not null`,
		id, userID,
	).Scan(
		&question.ID,
		&question.UserID,
		&question.ClientID,
		&question.QuestionCiphertext,
		&question.EncryptionVersion,
		&question.Answered,
		&question.ReturnAt,
		&question.CreatedAt,
		&question.UpdatedAt,
	)
	if err != nil {
		return entities.Question{}, fmt.Errorf("get question %d for user %d: %w", id, userID, err)
	}
	return question, nil
}

func (r *QuestionRepository) GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error) {
	rows, err := r.db.Query(
		ctx,
		`select id, user_id, client_id, question_ciphertext, encryption_version,
		        answered, coalesce(return_at, to_timestamp(0)), created_at, updated_at
		 from questions
		 where user_id = $1
		   and question_ciphertext is not null
		   and (return_at is null or return_at <= current_timestamp)
		 order by return_at, created_at desc`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query questions for user %d: %w", userID, err)
	}
	defer rows.Close()

	var questions []entities.Question
	for rows.Next() {
		var q entities.Question
		if err := rows.Scan(
			&q.ID,
			&q.UserID,
			&q.ClientID,
			&q.QuestionCiphertext,
			&q.EncryptionVersion,
			&q.Answered,
			&q.ReturnAt,
			&q.CreatedAt,
			&q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan question for user %d: %w", userID, err)
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questions for user %d: %w", userID, err)
	}
	return questions, nil
}

func (r *QuestionRepository) GetLegacyQuestions(ctx context.Context, userID int) ([]entities.LegacyQuestion, error) {
	rows, err := r.db.Query(
		ctx,
		`select id, question
		 from questions
		 where user_id = $1 and question_ciphertext is null and question is not null
		 order by id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query legacy questions for user %d: %w", userID, err)
	}
	defer rows.Close()

	var questions []entities.LegacyQuestion
	for rows.Next() {
		var q entities.LegacyQuestion
		if err := rows.Scan(&q.ID, &q.Question); err != nil {
			return nil, fmt.Errorf("scan legacy question for user %d: %w", userID, err)
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy questions for user %d: %w", userID, err)
	}
	return questions, nil
}

func (r *QuestionRepository) ImportLegacyQuestion(
	ctx context.Context,
	userID, id int,
	clientID uuid.UUID,
	ciphertext []byte,
	encryptionVersion int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`update questions
		 set client_id = $3,
		     question_ciphertext = $4,
		     encryption_version = $5,
		     question = null
		 where id = $1 and user_id = $2 and question_ciphertext is null`,
		id, userID, clientID, ciphertext, encryptionVersion,
	)
	if err != nil {
		return fmt.Errorf("import legacy question %d for user %d: %w", id, userID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("legacy question %d for user %d was not found or already migrated", id, userID)
	}
	return nil
}
