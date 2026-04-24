package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/models"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetAllUsers retrieves all users
func (h *UserHandler) GetAllUsers(c fiber.Ctx) error {
	logger.Info("开始获取所有用户")

	// Get all users from service
	users := h.userService.GetAllUsers()

	// Convert users to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	logger.Info("成功获取所有用户", zap.Int("count", len(userResponses)))
	return c.Status(fiber.StatusOK).JSON(userResponses)
}

type TransferAdminRequest struct {
	UserID string `json:"user_id"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
}

type ResetPasswordResponse struct {
	Message           string              `json:"message"`
	User              models.UserResponse `json:"user"`
	TemporaryPassword string              `json:"temporary_password"`
	RevokedSessions   int                 `json:"revoked_sessions"`
}

type DeleteUserResponse struct {
	Message             string              `json:"message"`
	User                models.UserResponse `json:"user"`
	TransferredNetworks int                 `json:"transferred_networks"`
	RevokedSessions     int                 `json:"revoked_sessions"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("创建用户失败：未认证")
		return authErr
	}

	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("创建用户失败，请求参数绑定失败", zap.String("current_user_id", currentUserID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	user, temporaryPassword, err := h.userService.CreateUserByAdmin(currentUserID, req.Username)
	if err != nil {
		logger.Error("创建用户失败", zap.String("current_user_id", currentUserID), zap.String("username", req.Username), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":            "用户创建成功，请通过其他方式安全告知用户临时密码",
		"user":               user.ToResponse(),
		"temporary_password": temporaryPassword,
	})
}

func (h *UserHandler) TransferAdmin(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("转让管理员失败：未认证")
		return authErr
	}
	logger.Info("开始转让管理员身份", zap.String("current_user_id", currentUserID))

	var req TransferAdminRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("转让管理员失败，请求参数绑定失败", zap.String("current_user_id", currentUserID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.userService.TransferAdmin(currentUserID, req.UserID)
	if err != nil {
		logger.Error("转让管理员失败", zap.String("current_user_id", currentUserID), zap.String("target_user_id", req.UserID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("成功转让管理员身份", zap.String("current_user_id", currentUserID), zap.String("target_user_id", req.UserID))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "管理员身份转让成功",
		"user":    user.ToResponse(),
	})
}

func (h *UserHandler) ResetPassword(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("重置用户密码失败：未认证")
		return authErr
	}
	targetUserID := c.Params("userId")

	logger.Info("开始重置用户密码",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID))

	user, temporaryPassword, revokedSessions, err := h.userService.ResetPasswordByAdmin(currentUserID, targetUserID)
	if err != nil {
		logger.Error("重置用户密码失败",
			zap.String("current_user_id", currentUserID),
			zap.String("target_user_id", targetUserID),
			zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("成功重置用户密码",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID),
		zap.Int("revoked_sessions", revokedSessions))

	return c.Status(fiber.StatusOK).JSON(ResetPasswordResponse{
		Message:           "密码已重置，请通过其他方式安全告知用户，并提醒其尽快修改密码",
		User:              user.ToResponse(),
		TemporaryPassword: temporaryPassword,
		RevokedSessions:   revokedSessions,
	})
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("删除用户失败：未认证")
		return authErr
	}
	targetUserID := c.Params("userId")

	logger.Info("开始删除用户",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID))

	user, transferredNetworks, revokedSessions, err := h.userService.DeleteUserByAdmin(currentUserID, targetUserID)
	if err != nil {
		logger.Error("删除用户失败",
			zap.String("current_user_id", currentUserID),
			zap.String("target_user_id", targetUserID),
			zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("成功删除用户",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID),
		zap.Int("transferred_networks", transferredNetworks),
		zap.Int("revoked_sessions", revokedSessions))

	return c.Status(fiber.StatusOK).JSON(DeleteUserResponse{
		Message:             "用户已删除，名下网络已转让给当前管理员",
		User:                user.ToResponse(),
		TransferredNetworks: transferredNetworks,
		RevokedSessions:     revokedSessions,
	})
}
