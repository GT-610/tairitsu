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

// UpdateUserRole updates a user's role
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

func (h *UserHandler) UpdateUserRole(c fiber.Ctx) error {
	userId := c.Params("userId")
	logger.Info("开始更新用户角色", zap.String("user_id", userId))

	// Bind and validate request body
	var req UpdateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("更新用户角色失败，请求参数绑定失败", zap.String("user_id", userId), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate role is either admin or user
	if req.Role != "admin" && req.Role != "user" {
		logger.Error("更新用户角色失败，角色值无效", zap.String("user_id", userId), zap.String("role", req.Role))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "无效的角色值，必须是admin或user"})
	}

	// Get user by ID
	user, err := h.userService.GetUserByID(userId)
	if err != nil {
		logger.Error("更新用户角色失败，获取用户时出错", zap.String("user_id", userId), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	// Update user role
	user.Role = req.Role

	// Save updated user to database
	if err := h.userService.UpdateUser(user); err != nil {
		logger.Error("更新用户角色失败，保存用户时出错", zap.String("user_id", userId), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "更新用户角色失败"})
	}

	logger.Info("成功更新用户角色", zap.String("user_id", userId), zap.String("new_role", req.Role))
	return c.Status(fiber.StatusOK).JSON(user.ToResponse())
}
