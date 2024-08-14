package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

// request errors
var errServer = errors.New("the server encountered a problem and could not process your request")
var errNotFound = errors.New("the requested resource could not be found")
var errMethodNotAllowed = func(r *http.Request) error {
	return fmt.Errorf("the %s method is not supported for this resource", r.Method)
}
var errWrongParameter = errors.New("the parameter is wrong")
var errMaxBytesRequest = func(maxBytesError *http.MaxBytesError) error {
	return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
}
var errRateLimit = errors.New("rate limit exceeded")

// json errors
var errJsonSyntax = func(syntaxError *json.SyntaxError) error {
	return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
}
var errJsonUnexpectedEOF = errors.New("body contains badly-formed JSON")
var errJsonUnmarshalType = func(unmarshalTypeError *json.UnmarshalTypeError) error {
	return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
}
var errJsonUnmarshalTypeField = func(unmarshalTypeError *json.UnmarshalTypeError) error {
	return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
}
var errJsonEOF = errors.New("body must not be empty")
var errJsonUnknownField = func(fieldName string) error {
	return fmt.Errorf("body contains unknown key %s", fieldName)
}
var errJsonSingleValue = errors.New("body must only contain a single JSON value")

// Log the error message and send response to the user with status code 500
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	err = app.errorLog.Output(2, trace)
	if err != nil {
		app.errorLog.Println(err.Error())
	}

	if app.config.env == "development" {
		app.clientError(w, trace, http.StatusInternalServerError)
	} else {
		app.clientError(w, errServer.Error(), http.StatusInternalServerError)
	}
}

// Send response to the user with the error info in json
func (app *application) clientError(w http.ResponseWriter, message any, status int) {

	env := wrapperJson{"error": message}

	err := app.writeJSON(w, env, status)
	if err != nil {
		app.errorLog.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
