package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invliad id parameter")
	}

	return id, nil
}

type envelope map[string]any

func (app *application) renderJSON(
	w http.ResponseWriter,
	status int,
	data envelope,
	headers http.Header,
) error {
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	ct := w.Header().Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.Split(ct, ";")[0])
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, message: msg}
		}
	}

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf(
				"request body contains badly-formed JSON (at position %d)",
				syntaxError.Offset,
			)
			return &malformedRequest{status: http.StatusBadRequest, message: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "request body contains badly-formed JSON"
			return &malformedRequest{status: http.StatusBadRequest, message: msg}

		case errors.As(err, &unmarshalTypeError):
			var msg string
			if unmarshalTypeError.Field != "" {
				msg = fmt.Sprintf(
					"request body contains incorrect JSON type for field %q",
					unmarshalTypeError.Field,
				)
			} else {
				msg = fmt.Sprintf(
					"request body contains incorrect JSON type (at character %d)",
					unmarshalTypeError.Offset,
				)
			}
			return &malformedRequest{status: http.StatusBadRequest, message: msg}

		case errors.Is(err, io.EOF):
			msg := "request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, message: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, message: msg}

		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("request body must not be larger than %d bytes", maxBytes)
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, message: msg}

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return &malformedRequest{status: http.StatusBadRequest, message: err.Error()}
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, message: msg}
	}

	return nil
}
