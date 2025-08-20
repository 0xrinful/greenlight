package main

import (
	"errors"
	"net/http"
	"time"

	"greenlight/internal/data"
	"greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.sendBadRequestError(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.sendServerError(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.sendValidationError(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.sendValidationError(w, r, v.Errors)
		default:
			app.sendServerError(w, r, err)
		}
		return
	}

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.sendServerError(w, r, err)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.sendServerError(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err := app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.renderJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.sendServerError(w, r, err)
	}
}

func (app *application) acivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"token"`
	}

	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.sendBadRequestError(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		app.sendValidationError(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expirey activation token")
			app.sendValidationError(w, r, v.Errors)
		default:
			app.sendServerError(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.sendEditConflictError(w, r)
		default:
			app.sendServerError(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.sendServerError(w, r, err)
		return
	}

	err = app.renderJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.sendServerError(w, r, err)
	}
}
