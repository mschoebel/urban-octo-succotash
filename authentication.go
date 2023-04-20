package uos

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

var cookieHandler *securecookie.SecureCookie

func setupAuthentication() {
	LogInfoContext(
		"initialize authentication",
		LogContext{
			"hash":  len(config.Auth.hash),
			"block": len(config.Auth.block),
		},
	)

	cookieHandler = securecookie.New(config.Auth.hash, config.Auth.block)
}

type sessionInfo struct {
	UserID     uint      `json:"id"`
	Expiration time.Time `json:"expiration"`
}

func setSession(userID uint, w http.ResponseWriter) {
	session := sessionInfo{
		UserID:     userID,
		Expiration: time.Now().Add(30 * time.Minute),
	}

	valueBytes, err := json.Marshal(session)
	if err != nil {
		LogErrorObj("could not encode session info as JSON", err)
		return
	}
	value := string(valueBytes)

	encoded, err := cookieHandler.Encode("session", value)
	if err != nil {
		LogErrorObj("could not encode session cookie", err)
		return
	}

	cookie := &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}
