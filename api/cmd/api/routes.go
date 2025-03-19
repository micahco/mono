package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/mono/lib/middleware"
)

// App router
func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Metrics)
	r.Use(app.recovery)
	r.Use(middleware.EnableCORS(app.config.cors.trustedOrigins))
	if app.config.limiter.enabled {
		r.Use(middleware.RateLimit(app.config.limiter.rps, app.errorResponse))
	}
	r.Use(app.authenticate)
	r.NotFound(app.handle(app.notFound))
	r.MethodNotAllowed(app.handle(app.methodNotAllowed))

	// Metrics
	r.Mount("/debug", middleware.Profiler())

	// API
	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.handle(app.healthcheck))

		r.Route("/tokens", func(r chi.Router) {
			r.Post("/authentication", app.handle(app.tokensAuthenticationPost))

			r.Route("/verification", func(r chi.Router) {
				r.Post("/registration", app.handle(app.tokensVerificaitonRegistrationPost))
				r.Post("/password-reset", app.handle(app.tokensVerificaitonPasswordResetPost))

				r.Route("/email-change", func(r chi.Router) {
					r.Use(app.requireAuthentication)

					r.Post("/", app.handle(app.tokensVerificaitonEmailChangePost))
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", app.handle(app.usersPost))
			r.Put("/password", app.handle(app.usersPasswordPut))

			r.Route("/me", func(r chi.Router) {
				r.Use(app.requireAuthentication)

				r.Get("/", app.handle(app.usersMeGet))
				r.Put("/", app.handle(app.usersMePut))
			})
		})
	})

	return r
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) error {
	return app.writeJSONError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) error {
	return app.writeJSONError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) error {
	env := "production"
	if app.config.dev {
		env = "development"
	}

	res := response{
		"status": "available",
		"system_info": map[string]string{
			"environment": env,
		},
	}

	return app.writeJSON(w, res, http.StatusOK)
}
