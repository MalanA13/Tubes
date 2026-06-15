// Package middleware menyediakan HTTP middleware untuk autentikasi dan otorisasi.
//
// Middleware menggunakan AuthClient (interface) sehingga mudah di-mock di test
// dan bisa diswap implementasinya tanpa mengubah kode middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/contextutil"
	"github.com/tubes-cc/logistics/internal/response"
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
			response.Unauthorized(w, "missing Authorization header")
			return
		}

		// Format harus "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(w, "invalid Authorization header format, use: Bearer <token>")
			return
		}

		token := parts[1]

		// Validasi token via AuthClient
		claims, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			response.Unauthorized(w, "token tidak valid: "+err.Error())
			return
		}

		// Simpan claims di context untuk digunakan handler
		ctx := context.WithValue(r.Context(), contextutil.AuthClaimsKey, claims)
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
		claims, ok := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
		if !ok || claims == nil {
			response.Unauthorized(w, "claims tidak ditemukan di context")
			return
		}

		if claims.Role != role {
			response.Forbidden(w, "akses ditolak: membutuhkan role "+string(role)+", user memiliki role "+string(claims.Role))
			return
		}

		next(w, r)
	}
}

// GetAuthClaims mengambil AuthClaims dari context request.
// Mengembalikan nil jika tidak ada (belum melewati middleware Authenticate).
func GetAuthClaims(r *http.Request) *domain.AuthClaims {
	claims, _ := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
	return claims
}
