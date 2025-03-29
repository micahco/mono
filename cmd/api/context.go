package main

import (
	"context"

	"github.com/micahco/mono/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(ctx context.Context, user *data.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func (app *application) contextGetUser(ctx context.Context) *data.User {
	user, ok := ctx.Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
