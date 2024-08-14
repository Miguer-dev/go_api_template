package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// create custom server with graceful shutdown. Start listen.
func (app *application) serve() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		ErrorLog:     app.errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	// background goruntime, waiting for shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)
		defer close(quit)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// block until a signal is received.
		s := <-quit

		app.infoLog.Printf("shutting down server, signal: %s", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		//20s context deadline for Shutdown
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.infoLog.Println("completing background tasks")
		app.wg.Wait()

		shutdownError <- nil
	}()

	app.infoLog.Printf("starting %s server on %s", app.config.env, srv.Addr)

	err := srv.ListenAndServe()
	// ErrServerClosed is a good Shutdown
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Shutdown with error
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.infoLog.Println("stopped server")

	return nil
}
