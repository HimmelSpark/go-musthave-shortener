package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{status: http.StatusOK}
		lw := &loggingResponseWriter{ResponseWriter: w, responseData: rd}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Dur("duration", duration).
			Int("status", rd.status).
			Int("size", rd.size).
			Msg("request completed")
	})
}
