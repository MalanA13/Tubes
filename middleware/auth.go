// Package middleware menyediakan HTTP middleware untuk autentikasi dan otorisasi.
//
// Middleware menggunakan AuthClient (interface) sehingga mudah di-mock di test
// dan bisa diswap implementasinya tanpa mengubah kode middleware.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/domain"
)

// contextKey adalah tipe untuk key dalam context agar menghindari collision.
type contextKey string

const (
	// ContextKeyAuthClaims adalah key untuk menyimpan AuthClaims di context.
	ContextKeyAuthClaims contextKey = "auth_claims"
)

// AuthMiddleware menyediakan middleware autentikasi berbasis JWT.
// Menggunakan AuthClient interface (bukan implementasi konkret).
type AuthMiddleware struct {
	authClient client.AuthClient
}

// NewAuthMiddleware membuat AuthMiddleware baru dengan dependency injection.
//
// Contoh penggunaan:
//
//	authClient := client.NewHTTPAuthClient("http://auth-service")
//	mw := middleware.NewAuthMiddleware(authClient)
//	http.Handle("/hub/scan-in", mw.Authenticate(hubHandler.ScanIn))
func NewAuthMiddleware(authClient client.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

// Authenticate adalah middleware yang memvalidasi token JWT dari header Authorization.
//
// Format header: "Authorization: Bearer <token>"
//
// Jika valid, AuthClaims disimpan di context dan handler berikutnya dipanggil.
// Jika tidak valid, mengembalikan 401 Unauthorized.
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSONError(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		// Format harus "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeJSONError(w, http.StatusUnauthorized, "invalid Authorization header format, use: Bearer <token>")
			return
		}

		token := parts[1]

		// Validasi token via AuthClient
		claims, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "token tidak valid: "+err.Error())
			return
		}

		// Simpan claims di context untuk digunakan handler
		ctx := context.WithValue(r.Context(), ContextKeyAuthClaims, claims)
		next(w, r.WithContext(ctx))
	}
}

// RequireRole adalah middleware yang memastikan user memiliki role tertentu.
// Harus digunakan SETELAH Authenticate.
//
// Contoh:
//
//	mw.Authenticate(mw.RequireRole(domain.RoleAdmin, adminHandler))
func (m *AuthMiddleware) RequireRole(role domain.UserRole, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ContextKeyAuthClaims).(*domain.AuthClaims)
		if !ok || claims == nil {
			writeJSONError(w, http.StatusUnauthorized, "claims tidak ditemukan di context")
			return
		}

		if claims.Role != role {
			writeJSONError(w, http.StatusForbidden,
				"akses ditolak: membutuhkan role "+string(role)+", user memiliki role "+string(claims.Role))
			return
		}

		next(w, r)
	}
}

// GetAuthClaims mengambil AuthClaims dari context request.
// Mengembalikan nil jika tidak ada (belum melewati middleware Authenticate).
func GetAuthClaims(r *http.Request) *domain.AuthClaims {
	claims, _ := r.Context().Value(ContextKeyAuthClaims).(*domain.AuthClaims)
	return claims
}

// middlewareErrorResponse adalah format response error dari middleware.
type middlewareErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeJSONError menulis response error dalam format JSON.
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(middlewareErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
