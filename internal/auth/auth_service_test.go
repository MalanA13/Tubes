package auth

import (
	"testing"
	models "github.com/tubes-cc/logistics/domain"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	// Setup Mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	// Tentukan Ekspektasi: 
	// "Kalau email 'tes@gmail.com' dipanggil, kasih balik data user ini ya!"
	dummyUser := &models.User{Email: "tes@gmail.com", Password: "123"}
	mockRepo.EXPECT().GetByEmail("tes@gmail.com").Return(dummyUser, nil)

	// Jalankan fungsi aslinya
	token, err := authService.Login("tes@gmail.com", "123")

	// Cek hasilnya
	assert.NoError(t, err)
	assert.Equal(t, "TOKEN_JWT_DUMMY", token)
}