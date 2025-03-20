package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/micahco/mono/web/ui"
)

func (app *application) render(w http.ResponseWriter, r *http.Request, statusCode int, title string, children templ.Component) error {
	w.WriteHeader(statusCode)

	ctx := templ.WithChildren(r.Context(), children)

	return ui.Base(title).Render(ctx, w)
}

func (app *application) renderError(w http.ResponseWriter, errorMessage string, statusCode int) error {
	http.Error(w, errorMessage, statusCode)

	return nil
}
