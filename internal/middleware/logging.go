package middleware

import (
	"net/http"
	"time"

	"github.com/tubes-cc/logistics/internal/contextutil"
	"go.uber.org/zap"
)

// responseWriter captures the HTTP status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs the incoming HTTP request and its duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK, // Default status if WriteHeader is not called
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		reqID := contextutil.GetRequestID(r.Context())

		zap.L().Info("HTTP request processed",
			zap.String("request_id", reqID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("duration", duration),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}
