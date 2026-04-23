package handlers

import (
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService    *services.UserService
	sessionService *services.SessionService
	jwtService     *services.JWTService
	runtimeService *services.RuntimeService
	stateService   *services.StateService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(userService *services.UserService, sessionService *services.SessionService, jwtService *services.JWTService, runtimeService *services.RuntimeService, stateService *services.StateService) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		sessionService: sessionService,
		jwtService:     jwtService,
		runtimeService: runtimeService,
		stateService:   stateService,
	}
}

// Register handles user registration requests
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("注册请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始用户注册", zap.String("username", req.Username))

	role := "user"
	if h.stateService != nil && h.stateService.IsInitialized() && !h.stateService.RuntimeSettings().AllowPublicRegistration {
		logger.Warn("公开注册已关闭，拒绝运行态注册", zap.String("username", req.Username))
		return writeUserServiceError(c, services.ErrPublicRegistrationDisabled)
	}
	if h.stateService != nil && !h.stateService.IsInitialized() {
		hasAdmin, err := h.userService.HasAdminUser()
		if err != nil {
			logger.Warn("检查管理员用户失败，使用默认角色", zap.Error(err))
		} else if !hasAdmin {
			role = "admin"
			logger.Info("系统尚未初始化且无管理员用户，将当前用户设为管理员", zap.String("username", req.Username))
		}
	}

	user, err := h.userService.Register(&req, role)
	if err != nil {
		logger.Error("用户注册失败", zap.String("username", req.Username), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("用户注册成功", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", user.Role))

	// Only the setup-time admin creation path should trigger ZeroTier client initialization.
	if h.runtimeService != nil && h.stateService != nil && !h.stateService.IsInitialized() {
		logger.Info("初始化ZeroTier客户端")
		if _, err := h.runtimeService.InitZTClientFromConfig(); err != nil {
			logger.Warn("ZeroTier客户端初始化失败，稍后将再次尝试", zap.Error(err))
		} else {
			logger.Info("ZeroTier客户端初始化和连接验证成功")
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
		return writeUserServiceError(c, err)
	}

	logger.Info("用户登录成功", zap.String("user_id", user.ID), zap.String("username", user.Username))

	session, err := h.sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  c.Get("User-Agent"),
		IPAddress:  c.IP(),
		RememberMe: req.RememberMe,
		ExpiresAt:  time.Now().Add(h.jwtService.AccessExpiry()),
	})
	if err != nil {
		logger.Error("创建登录会话失败", zap.String("user_id", user.ID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	// Generate JWT token for authenticated user
	token, err := h.jwtService.GenerateToken(user, session.ID)
	if err != nil {
		logger.Error("生成JWT令牌失败", zap.String("user_id", user.ID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "生成令牌失败"})
	}

	logger.Info("JWT令牌生成成功", zap.String("user_id", user.ID))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":   token,
		"user":    user.ToResponse(),
		"session": session.ToResponse(true),
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
		return writeUserServiceError(c, err)
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
	currentSessionID, _ := c.Locals("session_id").(string)

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

	revokedCount := 0
	if req.LogoutOtherSessions {
		count, err := h.userService.ChangePasswordAndRevokeOtherSessions(userID.(string), oldPassword, newPassword, currentSessionID)
		if err != nil {
			logger.Error("修改密码失败", zap.String("user_id", userID.(string)), zap.Error(err))
			return writeUserServiceError(c, err)
		}
		revokedCount = count
	} else if err := h.userService.ChangePassword(userID.(string), oldPassword, newPassword); err != nil {
		logger.Error("修改密码失败", zap.String("user_id", userID.(string)), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("密码修改成功", zap.String("user_id", userID.(string)))

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":                "密码修改成功",
		"revoked_other_sessions": revokedCount,
	})
}

// ListSessions returns the current user's active and historical sessions.
func (h *AuthHandler) ListSessions(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	currentSessionID, _ := c.Locals("session_id").(string)

	sessions, err := h.sessionService.GetUserSessions(userID)
	if err != nil {
		logger.Error("获取会话列表失败", zap.String("user_id", userID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	responses := make([]models.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		responses = append(responses, session.ToResponse(session.ID == currentSessionID))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"sessions": responses})
}

// Logout revokes the current session and clears the server-side login state.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	sessionID, _ := c.Locals("session_id").(string)

	if err := h.sessionService.RevokeSession(userID, sessionID); err != nil {
		logger.Error("退出登录失败", zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "已退出当前会话"})
}

// RevokeSession revokes one session owned by the current user.
func (h *AuthHandler) RevokeSession(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	sessionID := c.Params("sessionId")

	if err := h.sessionService.RevokeSession(userID, sessionID); err != nil {
		logger.Error("吊销会话失败", zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "会话已移除"})
}

// RevokeOtherSessions revokes all other sessions of the current user.
func (h *AuthHandler) RevokeOtherSessions(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	currentSessionID, _ := c.Locals("session_id").(string)

	count, err := h.sessionService.RevokeOtherSessions(userID, currentSessionID)
	if err != nil {
		logger.Error("移除其他会话失败", zap.String("user_id", userID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "其他会话已移除",
		"count":   count,
	})
}
