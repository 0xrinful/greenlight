package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.sendNotFoundError)
	router.MethodNotAllowed = http.HandlerFunc(app.sendMethodNotAllowedError)

	router.HandlerFunc("GET", "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(
		"POST", "/v1/movies",
		app.requirePermission("movies:write", app.createMovieHandler),
	)
	router.HandlerFunc(
		"GET", "/v1/movies",
		app.requirePermission("movies:read", app.listMoviesHandler),
	)
	router.HandlerFunc(
		"GET", "/v1/movies/:id",
		app.requirePermission("movies:read", app.showMovieHandler),
	)
	router.HandlerFunc(
		"PATCH", "/v1/movies/:id",
		app.requirePermission("movies:write", app.updateMovieHandler),
	)
	router.HandlerFunc(
		"DELETE", "/v1/movies/:id",
		app.requirePermission("movies:write", app.deleteMovieHandler),
	)

	router.HandlerFunc("POST", "/v1/users", app.registerUserHandler)
	router.HandlerFunc("PUT", "/v1/users/activated", app.acivateUserHandler)

	router.HandlerFunc("POST", "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler("GET", "/debug/vars", expvar.Handler())

	standard := alice.New(
		app.metrics, app.recoverPanic,
		app.enableCORS, app.rateLimit,
		app.authenticate,
	)

	return standard.Then(router)
}
