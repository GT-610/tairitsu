package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestUserServiceRegisterRequiresDatabase(t *testing.T) {
	userService := services.NewUserService(nil)

	user, err := userService.Register(&models.RegisterRequest{
		Username: "admin",
		Password: "secret123",
	})

	assert.Nil(t, user)
	assert.EqualError(t, err, "系统尚未配置数据库，请先完成初始设置")
}

func TestUserServiceLoginRequiresDatabase(t *testing.T) {
	userService := services.NewUserService(nil)

	user, err := userService.Login(&models.LoginRequest{
		Username: "admin",
		Password: "secret123",
	})

	assert.Nil(t, user)
	assert.EqualError(t, err, "系统尚未配置数据库，请先完成初始设置")
}
