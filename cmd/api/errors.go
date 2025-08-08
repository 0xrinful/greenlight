package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) sendError(w http.ResponseWriter, r *http.Request, status int, error any) {
	env := envelope{"error": error}

	err := app.renderJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) sendServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"

	app.sendError(w, r, http.StatusInternalServerError, message)
}

func (app *application) sendNotFoundError(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.sendError(w, r, http.StatusNotFound, message)
}

func (app *application) sendMethodNotAllowedError(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.sendError(w, r, http.StatusMethodNotAllowed, message)
}

type malformedRequest struct {
	status  int
	message string
}

func (e *malformedRequest) Error() string {
	return e.message
}

func (app *application) sendBadRequestError(w http.ResponseWriter, r *http.Request, err error) {
	var mr *malformedRequest
	if errors.As(err, &mr) {
		app.sendError(w, r, mr.status, mr.message)
		return
	}
	app.sendError(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) sendValidationError(
	w http.ResponseWriter,
	r *http.Request,
	errors map[string]string,
) {
	app.sendError(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) sendEditConflictError(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.sendError(w, r, http.StatusConflict, message)
}
