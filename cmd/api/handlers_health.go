package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.renderJSON(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}
