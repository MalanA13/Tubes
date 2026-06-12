package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tubes-cc/logistics/internal/contextutil"
)

// RequestIDMiddleware ensures each request has a unique X-Request-ID header
// and propagates it via the request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(contextutil.RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Inject into context
		ctx := contextutil.WithRequestID(r.Context(), reqID)

		// Set header in response
		w.Header().Set(contextutil.RequestIDHeader, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
