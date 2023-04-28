package uos

import (
	"net/http"
)

func mwWrap(h http.Handler, options AppRequestHandlerOptions) http.Handler {
	return mwContext(mwLogging(mwAuthentication(h, options.IsAuthRequired)))
}

func mwWrapF(f func(http.ResponseWriter, *http.Request), options AppRequestHandlerOptions) http.Handler {
	return mwWrap(http.HandlerFunc(f), options)
}
