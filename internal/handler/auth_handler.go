package handler

import (
    "net/http"
    "github.com/tubes-cc/logistics/internal/auth"
)

func HandleLogin(s auth.AuthService) http.HandlerFunc {
    // WAJIB return anonymous function seperti ini:
    return func(w http.ResponseWriter, r *http.Request) {
        // Jalankan logika login kamu di DALAM sini!
        // Contoh: token, err := s.Login(...)

        // Kirim respon HTTP, BUKAN "return token, nil" langsung ke luar.
        w.Write([]byte("Auth Service is Running!")) 
    }
}