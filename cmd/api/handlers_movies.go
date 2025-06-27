package main

import (
	"net/http"
	"time"

	"greenlight/internal/data"
)

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
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
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {}
