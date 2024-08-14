package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"

	"go.api.template/internal/models"
)

// Recover after panic, send a 500 status, this only work on the main goroutine, if you create a secundary goroutine, this will not work on it.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {

			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Limit requests for ip client.
func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while the cleanup is taking place.
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Retrieves the client IP address from any X-Forwarded-For or X-Real-IP headers, if neither of them are present use r.RemoteAddr
		ip := realip.FromRequest(r)

		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(
				rate.Limit(app.config.limiter.requestsPerSecond),
				app.config.limiter.bucket)}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.clientError(w, errRateLimit.Error(), http.StatusTooManyRequests)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)

	})
}

// authenticate the user if a Bearer token is given
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This indicates to any caches that the response may vary based on the value of the Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.clientError(w, models.ErrInvalidAuthenticationToken(w).Error(), http.StatusUnauthorized)
			return
		}

		token := headerParts[1]

		if v := models.ValidateTokenPlaintext(token); !v.Valid() {
			app.clientError(w, models.ErrInvalidAuthenticationToken(w).Error(), http.StatusUnauthorized)
			return
		}

		user, err := app.models.Users.GetForToken(models.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrRecordNotFound):
				app.clientError(w, models.ErrInvalidAuthenticationToken(w).Error(), http.StatusUnauthorized)
			default:
				app.serverError(w, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

// Verify that the user is authenticate
func (app *application) requireAuthenticatenUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.clientError(w, models.ErrAuthenticationRequired.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Verify that the user is active
func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if !user.Activated {
			app.clientError(w, models.ErrInactiveUser.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Verify that the user have an specific permission.
func (app *application) requirePermission(code string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}

		if !permissions.Include(code) {
			app.clientError(w, models.ErrNotPermitted.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Set header with CORS policy
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Access-Control-Request-Method")

		var ifPreflighRequest = func() {
			// treat it as a preflight request.
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if app.config.cors.setup == "all" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			ifPreflighRequest()
		} else {
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")

			if origin != "" {

				for _, value := range app.config.cors.whiteList {
					if origin == value {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						ifPreflighRequest()

						break
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// record request metrics
func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived            = expvar.NewInt("total_requests_received")
		totalResponsesSent               = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds  = expvar.NewInt("total_processing_time_μs")
		totalRequestsActive              = expvar.NewInt("total_requests_active")
		totalRequestsBetweenMetricsCalls = expvar.NewInt("total_requests_between_metrics_calls")

		totalResponsesSentByStatus = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)
		totalRequestsActive.Set(totalRequestsReceived.Value() - totalResponsesSent.Value())

		if r.URL.Path != "/debug/vars" {
			totalRequestsBetweenMetricsCalls.Add(1)
		}

		mw := &metricsResponseWriter{wrapped: w}

		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		if r.URL.Path == "/debug/vars" {
			totalRequestsBetweenMetricsCalls.Set(0)
		}

		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
