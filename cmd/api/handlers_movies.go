package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight/internal/data"
)

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.sendNotFoundError(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	err = app.renderJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.sendServerError(w, r, err)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int          `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.sendBadRequestError(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
