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

			Log.InfoContextR(
				r,
				"received request",
				LogContext{
					"time":   startTime,
					"method": r.Method,
					"url":    r.URL.Path,
				},
			)
			Metrics.GaugeInc(mRequestActive)
			Metrics.CounterInc(mRequestCount)

			next.ServeHTTP(lrw, r)
			duration := time.Since(startTime)

			Metrics.GaugeDec(mRequestActive)
			Metrics.CounterIncValueCondition(mRequestDuration, duration.Milliseconds(), lrw.statusCode < 500)
			Metrics.CounterIncCondition(mRequestFailed, lrw.statusCode >= 500)
			Metrics.CounterIncCondition(mRequestSlow, duration >= 2*time.Second)

			Log.InfoContextR(
				r,
				"done request",
				LogContext{
					"duration": duration,
					"method":   r.Method,
					"url":      r.URL.Path,
					"status":   lrw.statusCode,
					"slow":     duration >= 2*time.Second,
				},
			)
		},
	)
}
