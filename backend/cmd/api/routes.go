package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {

	router := httprouter.New()
	// NOTE: HanlderFunc() is an adapter to convert
	// normal functions into Handlers
	// it adds the serverhttp method to the function
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/hello", app.hello)
	// questions
	router.HandlerFunc(http.MethodGet, "/questions/:id", app.showQuestionHandler)
	router.HandlerFunc(http.MethodPost, "/questions", app.createQuestionHandler)
	router.HandlerFunc(http.MethodGet, "/questions", app.listQuestionsHandler)
	// answers
	router.HandlerFunc(http.MethodGet, "/answers", app.hello)
	router.HandlerFunc(http.MethodPost, "/answers", app.hello)

	return router
}
