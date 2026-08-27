package middlewares

import "net/http"

type Middleware func(http.Handler) http.Handler

func ChainMiddlewares(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
