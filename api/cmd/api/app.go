package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/micahco/mono/lib/data"
	"github.com/micahco/mono/lib/mailer"
)

var passwordLength = validation.Length(8, 72)

const (
	InvalidCredentailsMessage         = "invalid credentials"
	InvalidAuthenticationTokenMessage = "invalid or expired authentication token"
	AuthenticationRequiredMessage     = "you must be authenticated to access this resource"
	RateLimitExceededMessage          = "rate limit exceeded"
	defaultTimeout                    = 2 * time.Second
)

type envelope map[string]any

type application struct {
	config config
	logger *slog.Logger
	mailer *mailer.Mailer
	db     data.DB
	wg     sync.WaitGroup
}

func (app *application) serve(errLog *log.Logger) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     errLog,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		// Intercept signals
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", slog.String("signal", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", slog.String("addr", srv.Addr))

		// Block until WaitGroup is zero
		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", slog.String("addr", srv.Addr))

	err := srv.ListenAndServe()
	// http.ErrServerClosed is expected from srv.Shutdown()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", slog.String("addr", srv.Addr))

	return nil
}

func (app *application) background(fn func() error) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error("background process recovered from panic", slog.Any("err", err))
			}
		}()

		if err := fn(); err != nil {
			app.logger.Error("background process returned error", slog.Any("err", err))
		}
	}()
}

func (app *application) readJSON(r *http.Request, dst any) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, statusCode int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(js)

	return nil
}

func (app *application) writeError(w http.ResponseWriter, statusCode int, message any) error {
	if message == nil {
		message = http.StatusText(statusCode)
	}

	data := envelope{"error": message}

	return app.writeJSON(w, statusCode, data, nil)
}

func (app *application) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	data := envelope{"error": message}

	err := app.writeJSON(w, statusCode, data, nil)
	if err != nil {
		app.logger.Error("unable to write error response", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, logMsg string, err error) {
	app.logger.Error(logMsg, slog.Any("err", err), slog.String("type", fmt.Sprintf("%T", err)))

	app.errorResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	app.errorResponse(w, InvalidAuthenticationTokenMessage, http.StatusUnauthorized)
}
