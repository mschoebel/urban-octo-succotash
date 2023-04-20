package uos

import (
	"net/http"
)

func mwWrap(h http.Handler) http.Handler {
	return mwAuthentication(mwLogging(h))
}

func mwWrapF(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return mwWrap(http.HandlerFunc(f))
}
