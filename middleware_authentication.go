package uos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
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
				Log.DebugError("could not get session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			var sessionJSON string
			err = cookieHandler.Decode("session", cookie.Value, &sessionJSON)
			if err != nil {
				Log.WarnError("could not decode session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			var session sessionInfo
			err = json.Unmarshal([]byte(sessionJSON), &session)
			if err != nil {
				Log.WarnError("could not unmarshal session cookie", err)
				next.ServeHTTP(w, r)
				return
			}

			Log.DebugContext("initialized session context", LogContext{"userID": session.UserID})

			if time.Since(session.Expiration).Seconds() > 0 {
				// session expired -> continue without authentifiaction
				Log.DebugContext("session expired", LogContext{"expiration": session.Expiration})
				next.ServeHTTP(w, r)
				return
			}

			var user AppUser
			if session.UserID > 0 {
				err = DB.First(&user, session.UserID).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// invalid session (user not available) -> continue without authentification
					next.ServeHTTP(w, r)
					return
				}
				if err != nil {
					Log.ErrorObj("could not get app user", err)
					RespondInternalServerError(w)
					return
				}

				user.csrfToken = session.CSRFToken
			}

			ctx := context.WithValue(r.Context(), ctxAppUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
