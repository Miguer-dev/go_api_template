package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// Init httprouter, match routes + handlers + middlewares
func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.clientError(w, errNotFound.Error(), http.StatusNotFound)
	})
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.clientError(w, errMethodNotAllowed(r).Error(), http.StatusMethodNotAllowed)
	})

	user := alice.New(app.requireAuthenticatenUser, app.requireActivatedUser)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.Handler(http.MethodGet, "/v1/examples", user.Then(app.requirePermission("example:read", app.listExamplesHandler)))
	router.Handler(http.MethodPost, "/v1/examples", user.Then(app.requirePermission("example:write", app.createExampleHandler)))
	router.Handler(http.MethodGet, "/v1/example/:id", user.Then(app.requirePermission("example:read", app.showExampleHandler)))
	router.Handler(http.MethodPatch, "/v1/example/:id", user.Then(app.requirePermission("example:write", app.updateExampleHandler)))
	router.Handler(http.MethodDelete, "/v1/example/:id", user.Then(app.requirePermission("example:write", app.deleteExampleHandler)))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/authentication", app.authenticateUserHandler)

	// exposing metrics
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	standard := alice.New(app.metrics, app.recoverPanic, app.enableCORS, app.authenticate)

	if app.config.limiter.enabled {
		standard = standard.Append(app.rateLimit)
	}

	return standard.Then(router)

}
