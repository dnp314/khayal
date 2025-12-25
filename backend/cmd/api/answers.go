package main

import (
	"khayal/internal/data"
	"net/http"
	"time"
)

func (app *application) showAnswerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	answer, err := app.models.Answer.Get(id)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"answer": answer}, nil)
}

func (app *application) listAnswersHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	answers, err := app.models.Answer.GetAll(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	app.writeJSON(w, http.StatusOK, envelope{"answers": answers}, nil)
}

func (app *application) createAnswerHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Answer      string    `json:"answer"`
		QuestionID  int       `json:"question_id"`
		ScheduledAt time.Time `json:"scheduled_at"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	answer := &data.Answer{
		Answer:     input.Answer,
		QuestionID: input.QuestionID,
	}

	question := &data.Question{
		ScheduledAt: input.ScheduledAt,
	}

	err = app.models.Answer.Insert(answer, question)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"message": "Answer created successfully"}, nil)
}
