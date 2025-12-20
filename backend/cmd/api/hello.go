package main

import "net/http"

func (app *application) hello(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{"message": "hello"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
