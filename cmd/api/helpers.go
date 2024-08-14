package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.api.template/internal/validator"
)

// Get ID param from the request route
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, err
	}

	return id, nil
}

// Returns a string value from the query string, or the provided default value
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// Reads a string value from the query string and then splits it into a slice
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

// Reads a string value from the query string and converts it to an integer before returning.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

type wrapperJson map[string]any

// Build json and send in a response
func (app *application) writeJSON(w http.ResponseWriter, data wrapperJson, status int) error {

	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// read json from a request
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	// limit the size of the request body to 1MB, protection against denial of service attacks
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	// return an error if exists a field which cannot be mapped.
	dec.DisallowUnknownFields()

	// decode the firts json in the request body
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {

		case errors.As(err, &syntaxError):
			return errJsonSyntax(syntaxError)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errJsonUnexpectedEOF

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return errJsonUnmarshalTypeField(unmarshalTypeError)
			}
			return errJsonUnmarshalType(unmarshalTypeError)

		case errors.Is(err, io.EOF):
			return errJsonEOF

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return errJsonUnknownField(fieldName)

		case errors.As(err, &maxBytesError):
			return errMaxBytesRequest(maxBytesError)

		// pass an unsupported value to Decode()
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// if the r.body has more info return an error
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errJsonSingleValue
	}

	return nil
}

// execute function on the background with recover on Panic
func (app *application) backgroundFuncWithRecover(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		// catch any panic in this background function, and log an error message instead of terminating the application
		defer func() {
			if err := recover(); err != nil {
				app.errorLog.Println(err)
			}
		}()

		fn()
	}()
}
