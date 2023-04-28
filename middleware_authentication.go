package uos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

const (
	ctxAppUser string = "ctxAppUser"
)

var (
	authenticationPageURL string
)

func mwAuthentication(next http.Handler, doAuthRedirect bool) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			forwardWithoutSession := func() {
				clearSession(w)

				if !doAuthRedirect {
					// forward to next handler
					next.ServeHTTP(w, r)
					return
				}

				if authenticationPageURL == "" {
					Log.ErrorR(r, "auth page not specified - send 404")
					RespondNotFound(w)
					return
				}

				Log.TraceContextR(r, "redirect to auth page", LogContext{"URL": r.URL})
				http.Redirect(
					w, r,
					fmt.Sprintf("%s?p=%s", authenticationPageURL, r.URL),
					http.StatusFound,
				)
			}

			// read session cookie
			cookie, err := r.Cookie("session")
			if err != nil {
				Log.DebugErrorR(r, "could not get session cookie", err)
				forwardWithoutSession()
				return
			}

			var sessionJSON string
			err = cookieHandler.Decode("session", cookie.Value, &sessionJSON)
			if err != nil {
				Log.WarnErrorR(r, "could not decode session cookie", err)
				forwardWithoutSession()
				return
			}

			var session sessionInfo
			err = json.Unmarshal([]byte(sessionJSON), &session)
			if err != nil {
				Log.WarnErrorR(r, "could not unmarshal session cookie", err)
				forwardWithoutSession()
				return
			}

			Log.DebugContextR(r, "initialized session context", LogContext{"userID": session.UserID})

			if time.Since(session.Expiration).Seconds() > 0 {
				// session expired -> continue without authentifiaction
				Log.DebugContextR(r, "session expired", LogContext{"expiration": session.Expiration})
				forwardWithoutSession()
				return
			}

			// authentification page? -> redirect
			if r.URL.Path == authenticationPageURL {
				target := r.Form.Get("p")
				if target == "" {
					target = "/"
				}

				Log.TraceContextR(r, "already logged in - redirect", LogContext{"target": target})
				http.Redirect(w, r, target, http.StatusFound)
				return
			}

			var user AppUser
			if session.UserID > 0 {
				err = DB.First(&user, session.UserID).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// invalid session (user not available) -> continue without authentification
					forwardWithoutSession()
					return
				}
				if err != nil {
					Log.ErrorObjR(r, "could not get app user", err)
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
