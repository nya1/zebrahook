package zebrahook

import (
	"context"
	// "github.com/golang-jwt/jwt/v4"
)

type authInfo struct {
	// user   string
	userId string
	// claims jwt.MapClaims
	// key    string
}

type ctxValue int

const (
	ctxValueClaims ctxValue = iota
)

// contextWithAuthInfo adds the given JWT claims to the context and returns it.
func contextWithAuthInfo(ctx context.Context, auth authInfo) context.Context {
	return context.WithValue(ctx, ctxValueClaims, auth)
}

// contextAuthInfo returns the auth info from the context
func contextAuthInfo(ctx context.Context) (auth authInfo) {
	auth, _ = ctx.Value(ctxValueClaims).(authInfo)
	return
}
