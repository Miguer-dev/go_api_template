package main

import (
	"context"
	"net/http"

	"go.api.template/internal/models"
)

type contextKey string

const userContextKey = contextKey("user")

// save user info in context
func (app *application) contextSetUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// get user info from context
func (app *application) contextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
