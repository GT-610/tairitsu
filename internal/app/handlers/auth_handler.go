package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"github.com/tairitsu/tairitsu/internal/app/services"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(userService *services.UserService, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(\u0026req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(\u0026req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user.ToResponse(),
		"message": "注册成功",
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(\u0026req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Login(\u0026req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 生成令牌
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}