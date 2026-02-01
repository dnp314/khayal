package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()
	// NOTE: HanlderFunc() is an adapter to convert
	// normal functions into Handlers
	// it adds the serverhttp method to the function
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// questions
	router.HandlerFunc(http.MethodGet, "/questions/:id", app.requirePermission("questions:read", app.showQuestionHandler))
	router.HandlerFunc(http.MethodGet, "/questions", app.requirePermission("questions:read", app.listQuestionsHandler))
	router.HandlerFunc(http.MethodPost, "/questions", app.requirePermission("questions:write", app.createQuestionHandler))
	router.HandlerFunc(http.MethodDelete, "/questions/:id", app.requirePermission("questions:write", app.deleteQuestionHandler))
	// answers
	router.HandlerFunc(http.MethodGet, "/answers/:id", app.requireActivatedUser(app.showAnswerHandler))
	router.HandlerFunc(http.MethodGet, "/questions/:id/answers", app.requireActivatedUser(app.listAnswersHandler))
	router.HandlerFunc(http.MethodPost, "/answers", app.requireActivatedUser(app.createAnswerHandler))
	//users
	router.HandlerFunc(http.MethodPost, "/users/", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activated", app.activateUserHandler)
	//tokens
	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
