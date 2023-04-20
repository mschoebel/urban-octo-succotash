package uos

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	ctxAppUser string = "ctxAppUser"
)

func mwAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// read session cookie
			cookie, err := r.Cookie("session")
			if err != nil {
				LogDebugError("could not get session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			var sessionJSON string
			err = cookieHandler.Decode("session", cookie.Value, &sessionJSON)
			if err != nil {
				LogWarnError("could not decode session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			var session sessionInfo
			err = json.Unmarshal([]byte(sessionJSON), &session)
			if err != nil {
				LogWarnError("could not unmarshal session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			LogDebugContext("initialized session context", LogContext{"userID": session.UserID})

			var user AppUser
			if session.UserID > 0 {
				err = DB.First(&user, session.UserID).Error
				if err != nil {
					LogErrorObj("could not get app user", err)
					RespondInternalServerError(w)
					return
				}
			}

			ctx := context.WithValue(r.Context(), ctxAppUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}