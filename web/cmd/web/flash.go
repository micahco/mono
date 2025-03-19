package main

import (
	"net/http"

	"github.com/micahco/mono/web/internal/flash"
)

const (
	flashSessionKey = "flash"
)

func (app *application) putFlash(r *http.Request, f flash.Message) {
	app.sessionManager.Put(r.Context(), flashSessionKey, f)
}

func (app *application) popFlash(r *http.Request) *flash.Message {
	exists := app.sessionManager.Exists(r.Context(), flashSessionKey)
	if exists {
		f, ok := app.sessionManager.Pop(r.Context(), flashSessionKey).(flash.Message)
		if ok {
			return &f
		}
	}

	return nil
}
