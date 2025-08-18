package main

import (
	"errors"
	"net/http"
	"time"

	"greenlight/internal/data"
	"greenlight/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.sendBadRequestError(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.sendValidationError(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.sendInvalidCredentialsError(w, r)
		default:
			app.sendServerError(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.sendServerError(w, r, err)
		return

	}

	if !match {
		app.sendInvalidCredentialsError(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.sendServerError(w, r, err)
		return
	}

	err = app.renderJSON(w, http.StatusCreated, envelope{"token": token}, nil)
	if err != nil {
		app.sendServerError(w, r, err)
	}
}
