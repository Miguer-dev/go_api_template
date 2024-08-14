package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.api.template/internal/models"
	"go.api.template/internal/validator"
)

// Show API details
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := wrapperJson{
		"status": "available",
		"system_info": map[string]any{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, data, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
	}
}

// Create new example
func (app *application) createExampleHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ExampleValue1 float64 `json:"example_value_1"`
		ExampleValue2 string  `json:"example_value_2"`
		ExampleValue3 string  `json:"example_value_3"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.clientError(w, err.Error(), http.StatusBadRequest)
		return
	}

	example := &models.Example{
		ExampleValue1: input.ExampleValue1,
		ExampleValue2: input.ExampleValue2,
		ExampleValue3: input.ExampleValue3,
	}

	if v := example.ValidateExample(r); !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	err = app.models.Examples.Insert(example)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/v1/example/%d", example.Id))

	err = app.writeJSON(w, wrapperJson{"example": example}, http.StatusCreated)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// View example, search by id
func (app *application) showExampleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.clientError(w, errWrongParameter.Error(), http.StatusNotFound)
		return
	}

	data, err := app.models.Examples.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrExampleRecordNotFound) {
			app.clientError(w, err.Error(), http.StatusNotFound)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.writeJSON(w, wrapperJson{"example": data}, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// Update example
func (app *application) updateExampleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.clientError(w, errWrongParameter.Error(), http.StatusNotFound)
		return
	}

	var input struct {
		ExampleValue1 float64 `json:"example_value_1"`
		ExampleValue2 string  `json:"example_value_2"`
		ExampleValue3 string  `json:"example_value_3"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.clientError(w, err.Error(), http.StatusBadRequest)
		return
	}

	example := &models.Example{
		ExampleValue1: input.ExampleValue1,
		ExampleValue2: input.ExampleValue2,
		ExampleValue3: input.ExampleValue3,
	}

	if v := example.ValidateExample(r); !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	err = app.models.Examples.Update(example)
	if err != nil {
		if errors.Is(err, models.ErrExampleRecordNotFound) {
			app.clientError(w, err.Error(), http.StatusNotFound)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	w.Header().Set("Location", fmt.Sprintf("/v1/example/%d", example.Id))

	err = app.writeJSON(w, wrapperJson{"example": example}, http.StatusCreated)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// Delete example
func (app *application) deleteExampleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.clientError(w, errWrongParameter.Error(), http.StatusNotFound)
		return
	}

	err = app.models.Examples.Delete(id)
	if err != nil {
		if errors.Is(err, models.ErrExampleRecordNotFound) {
			app.clientError(w, err.Error(), http.StatusNotFound)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.writeJSON(w, wrapperJson{"message": fmt.Sprintf("example %d successfully deleted", id)}, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
		return
	}

}

// Gets examples, search by filters
func (app *application) listExamplesHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ExampleValue2 string
		ExampleValue3 string
		Filters       *models.Filters
	}

	v := &validator.Validator{}

	parameters := r.URL.Query()

	input.ExampleValue2 = app.readString(parameters, "example_value_2", "")
	input.ExampleValue3 = app.readString(parameters, "example_value_3", "")

	page := app.readInt(parameters, "page", 1, v)
	pageSize := app.readInt(parameters, "page_size", 20, v)
	sort := app.readString(parameters, "sort", "id")
	sortSafelist := []string{"id", "site"}

	input.Filters = models.InitFilters(page, pageSize, sort, sortSafelist)
	input.Filters.ValidateFilters(v)

	if !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	data, metadata, err := app.models.Examples.GetAll(input.ExampleValue2, input.ExampleValue3, input.Filters)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.writeJSON(w, wrapperJson{"metadata": metadata, "examples": data}, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// create a new user
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.clientError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &models.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if v := user.ValidateUser(); !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			app.clientError(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			app.serverError(w, err)
		}

		return
	}

	err = app.models.Permissions.AddForUser(user.ID, "example:read")
	if err != nil {
		app.serverError(w, err)
		return
	}

	token, err := app.models.Tokens.InitToken(user.ID, 3*24*time.Hour, models.ScopeActivation)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.backgroundFuncWithRecover(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.errorLog.Println(err.Error())
		}
	})

	err = app.writeJSON(w, wrapperJson{"user": user}, http.StatusCreated)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// activate an user
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.clientError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if v := models.ValidateTokenPlaintext(input.TokenPlaintext); !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	token := &models.Token{
		Plaintext: input.TokenPlaintext,
		Hash:      models.HashToken(input.TokenPlaintext),
		Scope:     models.ScopeActivation,
	}

	err = app.models.Tokens.GetActiveToken(token)
	if err != nil {
		if errors.Is(err, models.ErrTokenRecordNotFoundOrExpiry) {
			app.clientError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.models.Tokens.DeleteAllForUser(token.Scope, token.UserID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.models.Users.UpdateField(token.UserID, "activated", true)
	if err != nil {
		if errors.Is(err, models.ErrUserRecordNotFound) {
			app.clientError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	err = app.writeJSON(w, wrapperJson{"User": fmt.Sprintf("Id %d has been activated", token.UserID)}, http.StatusOK)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

// authenticate an user
func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email             string `json:"email"`
		PlaintextPassword string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.clientError(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := validator.Validator{}
	models.ValidateEmail(&v, input.Email)
	models.ValidatePassword(&v, input.PlaintextPassword)

	if !v.Valid() {
		app.clientError(w, v.Errors, http.StatusUnprocessableEntity)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserRecordNotFound) {
			app.clientError(w, models.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	if !user.Activated {
		app.clientError(w, models.ErrInactiveUser.Error(), http.StatusUnauthorized)
	}

	match, err := user.Password.Matches(input.PlaintextPassword)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !match {
		app.clientError(w, models.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		return
	}

	token, err := app.models.Tokens.InitToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.writeJSON(w, wrapperJson{"authentication_token": token}, http.StatusAccepted)
	if err != nil {
		app.serverError(w, err)
		return
	}
}
