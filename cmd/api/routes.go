package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.sendNotFoundError)
	router.MethodNotAllowed = http.HandlerFunc(app.sendMethodNotAllowedError)

	router.HandlerFunc("GET", "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc("POST", "/v1/movies", app.createMovieHandler)
	router.HandlerFunc("GET", "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc("GET", "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc("PATCH", "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc("DELETE", "/v1/movies/:id", app.deleteMovieHandler)

	router.HandlerFunc("POST", "/v1/users", app.registerUserHandler)

	standard := alice.New(app.recoverPanic, app.rateLimit)

	return standard.Then(router)
}
