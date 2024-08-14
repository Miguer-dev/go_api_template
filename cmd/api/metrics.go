package main

import (
	"database/sql"
	"expvar"
	"net/http"
	"runtime"
	"time"
)

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// get response headers
func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

// write response headers and save status code in the wrapper class
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

// write response and save status code 200 in the wrapper class
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !mw.headerWritten {
		mw.statusCode = http.StatusOK
		mw.headerWritten = true
	}

	return mw.wrapped.Write(b)
}

// get response
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

// init expvar variables for show server metrics
func initMetrics(db *sql.DB) {
	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}
