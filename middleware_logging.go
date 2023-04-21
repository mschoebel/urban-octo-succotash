package uos

import (
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func mwLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var (
				startTime = time.Now()
				lrw       = newLoggingResponseWriter(w)
			)

			Log.InfoContext(
				"received request",
				LogContext{
					"time":   startTime,
					"method": r.Method,
					"url":    r.URL.Path,
				},
			)
			next.ServeHTTP(lrw, r)
			Log.InfoContext(
				"done request",
				LogContext{
					"duration": time.Since(startTime),
					"method":   r.Method,
					"url":      r.URL.Path,
					"status":   lrw.statusCode,
				},
			)
		},
	)
}
