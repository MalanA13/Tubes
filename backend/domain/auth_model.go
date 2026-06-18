package domain

import "time"

type User struct {
    // ID menggunakan uint atau string (UUID)
    ID        uint      `json:"id" gorm:"primaryKey"`
    FullName  string    `json:"full_name"`
    Email     string    `json:"email" gorm:"unique"`
    Password  string    `json:"-"` // "-" artinya password tidak akan muncul saat data dikirim sebagai JSON
    Role      string    `json:"role"` // Contoh: "admin", "courier", "customer"
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Struct tambahan untuk request login dari user
type LoginRequest struct {
    Email    string `json:"email" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// Struct untuk response setelah login berhasil
type LoginResponse struct {
    Token string `json:"token"`
}

// Struct untuk request register akun baru
type RegisterRequest struct {
    FullName string `json:"full_name" binding:"required"`
    Email    string `json:"email" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// Struct untuk response setelah register berhasil
type RegisterResponse struct {
    ID        uint      `json:"id"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}