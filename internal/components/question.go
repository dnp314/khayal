package components

import (
	"time"

	"github.com/dnp314/khayal/internal/entities"
)

// canAnswerQuestion reports whether the question is currently due. A question
// with no return time, or one whose return time has passed, can be answered.
func canAnswerQuestion(question entities.Question) bool {
	return question.ReturnAt == nil || !question.ReturnAt.After(time.Now())
}
