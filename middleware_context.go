package uos

import (
	"context"
	"net/http"
	"strings"
)

const (
	ctxRequestID      string = "ctxRequestID"
	ctxClientLanguage string = "ctxClientLanguage"
)

func mwContext(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// setup context for request handling
			// .. create random request ID
			ctx := context.WithValue(r.Context(), ctxRequestID, randomString(8))

			// .. get client language
			language := strings.Split(r.Header.Get("Accept-Language"), ",")[0]

			cookie, err := r.Cookie("language")
			if err != nil {
				Log.DebugErrorR(r, "could not get language cookie", err)
			} else {
				language = cookie.Value
			}
			ctx = context.WithValue(ctx, ctxClientLanguage, language)

			r = r.WithContext(ctx)

			//  parse URL form data (might be empty)
			err = r.ParseForm()
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
