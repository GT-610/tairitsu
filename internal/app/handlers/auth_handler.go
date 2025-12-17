package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(userService *services.UserService, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

// Register handles user registration requests
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("注册请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始用户注册", zap.String("username", req.Username), zap.String("email", req.Email))

	// Determine user role - first registered user becomes admin
	role := "user"
	hasAdmin, err := h.userService.HasAdminUser()
	if err != nil {
		// If admin check fails, log warning but continue with default role
		logger.Warn("检查管理员用户失败，使用默认角色", zap.Error(err))
	} else if !hasAdmin {
		// If no admin exists, set current user as admin
		role = "admin"
		logger.Info("系统中无管理员用户，将当前用户设为管理员", zap.String("username", req.Username))
	}

	user, err := h.userService.Register(&req, role)
	if err != nil {
		logger.Error("用户注册失败", zap.String("username", req.Username), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", user.Role))

	// Initialize ZeroTier client after successful registration
	// This ensures the client is instantiated before any ZeroTier operations
	logger.Info("初始化ZeroTier客户端")
	ztClient, err := zerotier.NewClient()
	if err != nil {
		// Log error but don't block registration
		logger.Warn("ZeroTier客户端初始化失败，稍后将再次尝试", zap.Error(err))
	} else {
		// Verify connection if client initialized successfully
		_, err = ztClient.GetStatus()
		if err != nil {
			logger.Warn("ZeroTier连接验证失败，稍后将再次尝试", zap.Error(err))
		} else {
			logger.Info("ZeroTier客户端初始化和连接验证成功")
			// Connection verification only - client will be created on-demand when needed
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":    user.ToResponse(),
		"message": "注册成功",
	})
}

// Login handles user authentication requests
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("登录请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("用户尝试登录", zap.String("username", req.Username))

	user, err := h.userService.Login(&req)
	if err != nil {
		logger.Error("用户登录失败", zap.String("username", req.Username), zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	// Generate JWT token for authenticated user
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("生成JWT令牌失败", zap.String("user_id", user.ID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "生成令牌失败"})
	}

	logger.Info("JWT令牌生成成功", zap.String("user_id", user.ID))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// GetProfile retrieves the authenticated user's profile information
func (h *AuthHandler) GetProfile(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户信息失败：未认证")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未认证"})
	}

	logger.Info("获取用户信息", zap.String("user_id", userID.(string)))

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		logger.Error("获取用户信息失败", zap.String("user_id", userID.(string)), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("获取用户信息成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return c.Status(fiber.StatusOK).JSON(user.ToResponse())
}

// ChangePassword handles user password change requests
func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("修改密码失败：未认证")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未认证"})
	}

	// Bind request body
	var req models.ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("修改密码请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始处理修改密码请求", zap.String("user_id", userID.(string)))

	// Determine which password fields to use (support both old and new formats)
	oldPassword := req.OldPassword
	newPassword := req.NewPassword

	// If new format fields are present, use them
	if req.CurrentPassword != "" {
		oldPassword = req.CurrentPassword
	}
	if req.NewPasswordField != "" {
		newPassword = req.NewPasswordField
	}

	// Validate new password and confirm password match if using new format
	if req.ConfirmPassword != "" && req.NewPasswordField != "" && req.NewPasswordField != req.ConfirmPassword {
		logger.Error("修改密码失败：新密码与确认密码不匹配", zap.String("user_id", userID.(string)))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "新密码与确认密码不匹配"})
	}

	// Call user service to change password
	if err := h.userService.ChangePassword(userID.(string), oldPassword, newPassword); err != nil {
		logger.Error("修改密码失败", zap.String("user_id", userID.(string)), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("密码修改成功", zap.String("user_id", userID.(string)))

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "密码修改成功"})
}
