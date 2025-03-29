package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"
	"github.com/micahco/mono/lib/middleware"
	"github.com/micahco/mono/web/assets"
	"github.com/micahco/mono/web/pages"
)

// App router
func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(app.recovery)
	r.Use(middleware.SecureHeaders)

	// Static files
	r.Handle("/static/*", app.handleStatic())
	r.Get("/favicon.ico", app.handleFavicon)

	r.Route("/", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(middleware.NoSurf(app.csrfFailureHandler()))
		r.Use(app.authenticate)

		r.Get("/", app.handle(app.getIndex))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", app.handle(app.handleAuthLoginPost))
			r.Post("/logout", app.handle(app.handleAuthLogoutPost))
			r.Post("/signup", app.handle(app.handleAuthSignupPost))
			r.Get("/register", app.handle(app.handleAuthRegisterGet))
			r.Post("/register", app.handle(app.handleAuthRegisterPost))
			r.Get("/reset", app.handle(app.handleAuthResetGet))
			r.Post("/reset", app.handle(app.handleAuthResetPost))
			r.Get("/reset/update", app.handle(app.handleAuthResetUpdateGet))
			r.Post("/reset/update", app.handle(app.handleAuthResetUpdatePost))
		})
	})

	return r
}

func (app *application) refresh(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (app *application) handleStatic() http.Handler {
	return http.FileServer(http.FS(assets.StaticFiles))
}

func (app *application) handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, assets.StaticFiles, "static/favicon.ico")
}

func (app *application) getIndex(w http.ResponseWriter, r *http.Request) error {
	if app.isAuthenticated(r) {
		suid, err := app.getSessionUserID(r)
		if err != nil {
			return err
		}

		user, err := app.db.Users.Get(r.Context(), suid)
		if err != nil {
			return err
		}

		component := pages.Dashboard(user)
		return app.render(w, r, http.StatusOK, "Dashboard", component)
	}

	component := pages.Login(nosurf.Token(r), app.popFormErrors(r))
	return app.render(w, r, http.StatusOK, "Welcome", component)
}
