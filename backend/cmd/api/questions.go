package main

import (
	"database/sql"
	"errors"
	"fmt"
	"khayal/internal/data"
	"khayal/internal/validator"
	"net/http"
	"strings"
)

func (app *application) showQuestionHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	question, err := app.models.Questions.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"question": question}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createQuestionHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title       string
		Description sql.NullString
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	question := &data.Question{
		Title:       input.Title,
		Description: input.Description,
	}

	// TODO:adding validation

	err = app.models.Questions.Insert(question)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/questions/%d", question.ID))

	app.writeJSON(w, http.StatusOK, envelope{"message": "question created successfully"}, headers)
}

func (app *application) listQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Title = strings.TrimSpace(input.Title)
	input.Genres = app.readCSV(qs, "genres", nil)
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.Pagesize = app.readInt(qs, "pagesize", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := app.models.Questions.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// is pointer aware
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Questions.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
