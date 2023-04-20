package uos

import (
	"net/http"
)

type logoutAction struct{}

func (a logoutAction) Name() string {
	return "logout"
}

func (a logoutAction) Do(w http.ResponseWriter, r *http.Request) *ResponseAction {
	return ResponseClearSessionCookie()
}
