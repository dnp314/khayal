package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Questions QuestionModel
	Answers   AnswerModel
	Users     UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Questions: QuestionModel{DB: db},
		Answers:   AnswerModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
