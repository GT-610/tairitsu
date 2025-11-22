package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tairitsu/tairitsu/internal/app/logger"
	"github.com/tairitsu/tairitsu/internal/app/models"
	"github.com/tairitsu/tairitsu/internal/app/services"
	"github.com/tairitsu/tairitsu/internal/zerotier"
	"go.uber.org/zap"
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
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("注册请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("开始用户注册", zap.String("username", req.Username), zap.String("email", req.Email))

	user, err := h.userService.Register(&req)
	if err != nil {
		logger.Error("用户注册失败", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	// 在用户注册成功后初始化ZeroTier客户端
	// 这将确保在进入ZeroTier检查环节前，客户端已完成实例化和准备工作
	logger.Info("初始化ZeroTier客户端")
	ztClient, err := zerotier.NewClient()
	if err != nil {
		// 记录错误但不阻止用户注册流程
		logger.Warn("ZeroTier客户端初始化失败，稍后将再次尝试", zap.Error(err))
	} else {
		// 如果客户端初始化成功，验证连接
		_, err = ztClient.GetStatus()
		if err != nil {
			logger.Warn("ZeroTier连接验证失败，稍后将再次尝试", zap.Error(err))
		} else {
			logger.Info("ZeroTier客户端初始化和连接验证成功")
			// 存储到全局状态或缓存中，供后续使用
			// 这里只是验证连接，实际使用时会按需创建客户端
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user.ToResponse(),
		"message": "注册成功",
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("登录请求参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("用户尝试登录", zap.String("username", req.Username))

	user, err := h.userService.Login(&req)
	if err != nil {
		logger.Error("用户登录失败", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	// 生成令牌
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("生成JWT令牌失败", zap.String("user_id", user.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}
	
	logger.Info("JWT令牌生成成功", zap.String("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("获取用户信息失败：未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	
	logger.Info("获取用户信息", zap.String("user_id", userID.(string)))

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		logger.Error("获取用户信息失败", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	logger.Info("获取用户信息成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	c.JSON(http.StatusOK, user.ToResponse())
}