package uos

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

var cookieHandler *securecookie.SecureCookie

func setupAuthentication() {
	Log.InfoContext(
		"initialize authentication",
		LogContext{
			"hash":  len(Config.Auth.hash),
			"block": len(Config.Auth.block),
		},
	)

	cookieHandler = securecookie.New(Config.Auth.hash, Config.Auth.block)
}

type sessionInfo struct {
	UserID     uint      `json:"id"`
	Expiration time.Time `json:"expiration"`
	CSRFToken  string    `json:"token"`
}

func setSession(userID uint, w http.ResponseWriter) {
	session := sessionInfo{
		UserID:     userID,
		Expiration: time.Now().Add(30 * time.Minute),
		CSRFToken:  randomString(32),
	}

	valueBytes, err := json.Marshal(session)
	if err != nil {
		Log.ErrorObj("could not encode session info as JSON", err)
		return
	}
	value := string(valueBytes)

	encoded, err := cookieHandler.Encode("session", value)
	if err != nil {
		Log.ErrorObj("could not encode session cookie", err)
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

func setLanguage(language string, w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:  "language",
		Value: language,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}
