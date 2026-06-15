package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/tubes-cc/logistics/internal/contextutil"
	"github.com/tubes-cc/logistics/internal/response"
	"go.uber.org/zap"
)

// RecoveryMiddleware catches panics, logs them with Request-ID and stack trace,
// and returns a standardized JSON 500 error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				reqID := contextutil.GetRequestID(r.Context())
				
				// Log the panic with the required fields: request_id, panic value (error), stack trace
				zap.L().Error("Request panicked",
					zap.String("request_id", reqID),
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
				
				// Return standardized JSON 500 response using existing response package
				response.InternalServerError(w, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
