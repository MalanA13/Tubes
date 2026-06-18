package contextutil

type ContextKey string

const (
	RequestIDKey  ContextKey = "request_id"
	AuthClaimsKey ContextKey = "auth_claims"
)
