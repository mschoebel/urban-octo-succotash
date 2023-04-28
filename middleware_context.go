package uos

import (
	"context"
	"net/http"
)

const (
	ctxRequestID string = "ctxRequestID"
)

func mwContext(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// setup context for request handling
			// .. create random request ID
			ctx := context.WithValue(r.Context(), ctxRequestID, randomString(8))

			r = r.WithContext(ctx)

			//  parse URL form data (might be empty)
			err := r.ParseForm()
			if err != nil {
				Log.WarnErrorR(r, "could not parse form", err)
				RespondBadRequest(w)
				return
			}

			// forward to next handler
			next.ServeHTTP(w, r)
		},
	)
}
