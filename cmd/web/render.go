package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/justinas/nosurf"
	"github.com/micahco/mono/ui/pages"
)

func (app *application) render(w http.ResponseWriter, r *http.Request, statusCode int, title string, component templ.Component) error {
	w.WriteHeader(statusCode)

	page := pages.Base(title, nosurf.Token(r), app.isAuthenticated(r))
	ctx := templ.WithChildren(r.Context(), component)

	return page.Render(ctx, w)
}

func (app *application) renderError(w http.ResponseWriter, errorMessage string, statusCode int) error {
	http.Error(w, errorMessage, statusCode)

	return nil
}
