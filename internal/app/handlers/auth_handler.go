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
		logger.Error("Failed to bind registration request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Starting user registration", zap.String("username", req.Username))

	role := "user"
	if h.stateService != nil && h.stateService.IsInitialized() && !h.stateService.RuntimeSettings().AllowPublicRegistration {
		logger.Warn("Public registration is disabled; rejecting runtime registration", zap.String("username", req.Username))
		return writeUserServiceError(c, services.ErrPublicRegistrationDisabled)
	}
	if h.stateService != nil && !h.stateService.IsInitialized() {
		hasAdmin, err := h.userService.HasAdminUser()
		if err != nil {
			logger.Warn("Failed to check administrator users; using default role", zap.Error(err))
		} else if !hasAdmin {
			role = "admin"
			logger.Info("System is not initialized and no administrator exists; assigning administrator role", zap.String("username", req.Username))
		}
	}

	user, err := h.userService.Register(&req, role)
	if err != nil {
		logger.Error("User registration failed", zap.String("username", req.Username), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("User registered successfully", zap.String("user_id", user.ID), zap.String("username", user.Username), zap.String("role", user.Role))

	// Only the setup-time admin creation path should trigger ZeroTier client initialization.
	if h.runtimeService != nil && h.stateService != nil && !h.stateService.IsInitialized() {
		logger.Info("Initializing ZeroTier client")
		if _, err := h.runtimeService.InitZTClientFromConfig(); err != nil {
			logger.Warn("ZeroTier client initialization failed; it will be retried later", zap.Error(err))
		} else {
			logger.Info("ZeroTier client initialized and connection validated")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":         user.ToResponse(),
		"message":      "Registration succeeded",
		"message_code": "auth.registration_success",
	})
}

// Login handles user authentication requests
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind login request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("User login attempt", zap.String("username", req.Username))

	user, err := h.userService.Login(&req)
	if err != nil {
		logger.Error("User login failed", zap.String("username", req.Username), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("User logged in successfully", zap.String("user_id", user.ID), zap.String("username", user.Username))

	session, err := h.sessionService.CreateSession(services.SessionCreateInput{
		UserID:     user.ID,
		UserAgent:  c.Get("User-Agent"),
		IPAddress:  c.IP(),
		RememberMe: req.RememberMe,
		ExpiresAt:  time.Now().Add(h.jwtService.AccessExpiry()),
	})
	if err != nil {
		logger.Error("Failed to create login session", zap.String("user_id", user.ID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	// Generate JWT token for authenticated user
	token, err := h.jwtService.GenerateToken(user, session.ID)
	if err != nil {
		logger.Error("Failed to generate JWT token", zap.String("user_id", user.ID), zap.Error(err))
		return writeErrorResponseWithCode(c, fiber.StatusInternalServerError, "auth.token_generation_failed", "Failed to generate token")
	}

	logger.Info("JWT token generated successfully", zap.String("user_id", user.ID))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":   token,
		"user":    user.ToResponse(),
		"session": session.ToResponse(true),
	})
}

// GetProfile retrieves the authenticated user's profile information
func (h *AuthHandler) GetProfile(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user profile: unauthenticated")
		return authErr
	}

	logger.Info("Getting user profile", zap.String("user_id", userID))

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		logger.Error("Failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("User profile retrieved successfully", zap.String("user_id", user.ID), zap.String("username", user.Username))

	return c.Status(fiber.StatusOK).JSON(user.ToResponse())
}

// ChangePassword handles user password change requests
func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to change password: unauthenticated")
		return authErr
	}

	// Bind request body
	var req models.ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind change password request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Processing password change request", zap.String("user_id", userID))
	currentSessionID, _ := c.Locals("session_id").(string)

	if req.NewPassword != req.ConfirmPassword {
		logger.Error("Password change failed: confirmation mismatch", zap.String("user_id", userID))
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "auth.password_confirmation_mismatch", "The new password and confirmation do not match")
	}

	revokedCount := 0
	if req.LogoutOtherSessions {
		count, err := h.userService.ChangePasswordAndRevokeOtherSessions(userID, req.CurrentPassword, req.NewPassword, currentSessionID)
		if err != nil {
			logger.Error("Password change failed", zap.String("user_id", userID), zap.Error(err))
			return writeUserServiceError(c, err)
		}
		revokedCount = count
	} else if err := h.userService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		logger.Error("Password change failed", zap.String("user_id", userID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("Password changed successfully", zap.String("user_id", userID))

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":                "Password updated successfully",
		"message_code":           "auth.password_updated",
		"revoked_other_sessions": revokedCount,
	})
}

// ListSessions returns the current user's active and historical sessions.
func (h *AuthHandler) ListSessions(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to list sessions: unauthenticated")
		return authErr
	}
	currentSessionID, _ := c.Locals("session_id").(string)

	sessions, err := h.sessionService.GetUserSessions(userID)
	if err != nil {
		logger.Error("Failed to list sessions", zap.String("user_id", userID), zap.Error(err))
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
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Logout failed: unauthenticated")
		return authErr
	}
	sessionID, _ := c.Locals("session_id").(string)

	if err := h.sessionService.RevokeSession(userID, sessionID); err != nil {
		logger.Error("Logout failed", zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return writeMessageResponse(c, fiber.StatusOK, "auth.logout_success", "Current session signed out", nil)
}

// RevokeSession revokes one session owned by the current user.
func (h *AuthHandler) RevokeSession(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to revoke session: unauthenticated")
		return authErr
	}
	sessionID := c.Params("sessionId")

	if err := h.sessionService.RevokeSession(userID, sessionID); err != nil {
		logger.Error("Failed to revoke session", zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return writeMessageResponse(c, fiber.StatusOK, "auth.session_removed", "Session removed", nil)
}

// RevokeOtherSessions revokes all other sessions of the current user.
func (h *AuthHandler) RevokeOtherSessions(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to revoke other sessions: unauthenticated")
		return authErr
	}
	currentSessionID, _ := c.Locals("session_id").(string)

	count, err := h.sessionService.RevokeOtherSessions(userID, currentSessionID)
	if err != nil {
		logger.Error("Failed to revoke other sessions", zap.String("user_id", userID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Other sessions removed",
		"message_code": "auth.other_sessions_removed",
		"count":        count,
	})
}
