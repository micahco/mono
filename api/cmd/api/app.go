package main

import (
	"context"
	"errors"
	"fmt"
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
