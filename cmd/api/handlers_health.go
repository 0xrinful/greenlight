package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.renderJSON(w, http.StatusOK, env, nil)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}
