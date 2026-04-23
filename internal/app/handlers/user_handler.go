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

func (h *UserHandler) TransferAdmin(c fiber.Ctx) error {
	currentUserID, _ := c.Locals("user_id").(string)
	logger.Info("开始转让管理员身份", zap.String("current_user_id", currentUserID))

	var req TransferAdminRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("转让管理员失败，请求参数绑定失败", zap.String("current_user_id", currentUserID), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
