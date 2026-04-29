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
	logger.Info("Getting all users")

	// Get all users from service
	users := h.userService.GetAllUsers()

	// Convert users to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	logger.Info("All users retrieved", zap.Int("count", len(userResponses)))
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
	MessageCode       string              `json:"message_code"`
	User              models.UserResponse `json:"user"`
	TemporaryPassword string              `json:"temporary_password"`
	RevokedSessions   int                 `json:"revoked_sessions"`
}

type DeleteUserResponse struct {
	Message             string              `json:"message"`
	MessageCode         string              `json:"message_code"`
	User                models.UserResponse `json:"user"`
	TransferredNetworks int                 `json:"transferred_networks"`
	RevokedSessions     int                 `json:"revoked_sessions"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to create user: unauthenticated")
		return authErr
	}

	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to create user: request binding failed", zap.String("current_user_id", currentUserID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	user, temporaryPassword, err := h.userService.CreateUserByAdmin(currentUserID, req.Username)
	if err != nil {
		logger.Error("Failed to create user", zap.String("current_user_id", currentUserID), zap.String("username", req.Username), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":            "User created successfully. Share the temporary password securely outside this system.",
		"message_code":       "user.created",
		"user":               user.ToResponse(),
		"temporary_password": temporaryPassword,
	})
}

func (h *UserHandler) TransferAdmin(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to transfer administrator role: unauthenticated")
		return authErr
	}
	logger.Info("Transferring administrator role", zap.String("current_user_id", currentUserID))

	var req TransferAdminRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to transfer administrator role: request binding failed", zap.String("current_user_id", currentUserID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.userService.TransferAdmin(currentUserID, req.UserID)
	if err != nil {
		logger.Error("Failed to transfer administrator role", zap.String("current_user_id", currentUserID), zap.String("target_user_id", req.UserID), zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("Administrator role transferred", zap.String("current_user_id", currentUserID), zap.String("target_user_id", req.UserID))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Administrator role transferred successfully",
		"message_code": "user.admin_transferred",
		"user":         user.ToResponse(),
	})
}

func (h *UserHandler) ResetPassword(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to reset user password: unauthenticated")
		return authErr
	}
	targetUserID := c.Params("userId")

	logger.Info("Resetting user password",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID))

	user, temporaryPassword, revokedSessions, err := h.userService.ResetPasswordByAdmin(currentUserID, targetUserID)
	if err != nil {
		logger.Error("Failed to reset user password",
			zap.String("current_user_id", currentUserID),
			zap.String("target_user_id", targetUserID),
			zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("User password reset successfully",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID),
		zap.Int("revoked_sessions", revokedSessions))

	return c.Status(fiber.StatusOK).JSON(ResetPasswordResponse{
		Message:           "Password reset successfully. Share it securely outside this system and ask the user to change it soon.",
		MessageCode:       "user.password_reset",
		User:              user.ToResponse(),
		TemporaryPassword: temporaryPassword,
		RevokedSessions:   revokedSessions,
	})
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	currentUserID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to delete user: unauthenticated")
		return authErr
	}
	targetUserID := c.Params("userId")

	logger.Info("Deleting user",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID))

	user, transferredNetworks, revokedSessions, err := h.userService.DeleteUserByAdmin(currentUserID, targetUserID)
	if err != nil {
		logger.Error("Failed to delete user",
			zap.String("current_user_id", currentUserID),
			zap.String("target_user_id", targetUserID),
			zap.Error(err))
		return writeUserServiceError(c, err)
	}

	logger.Info("User deleted successfully",
		zap.String("current_user_id", currentUserID),
		zap.String("target_user_id", targetUserID),
		zap.Int("transferred_networks", transferredNetworks),
		zap.Int("revoked_sessions", revokedSessions))

	return c.Status(fiber.StatusOK).JSON(DeleteUserResponse{
		Message:             "User deleted. Their networks were transferred to the current administrator.",
		MessageCode:         "user.deleted",
		User:                user.ToResponse(),
		TransferredNetworks: transferredNetworks,
		RevokedSessions:     revokedSessions,
	})
}
