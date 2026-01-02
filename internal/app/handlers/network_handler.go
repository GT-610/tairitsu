package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// NetworkHandler 网络处理器
type NetworkHandler struct {
	networkService *services.NetworkService
}

// NewNetworkHandler 创建网络处理器实例
func NewNetworkHandler(networkService *services.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// GetStatus 获取ZeroTier网络状态
func (h *NetworkHandler) GetStatus(c fiber.Ctx) error {
	logger.Info("开始获取ZeroTier网络状态")

	status, err := h.networkService.GetStatus()
	if err != nil {
		logger.Error("获取ZeroTier网络状态失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功获取ZeroTier网络状态")

	return c.Status(fiber.StatusOK).JSON(status)
}

// GetNetworks 获取当前用户的所有网络
func (h *NetworkHandler) GetNetworks(c fiber.Ctx) error {
	logger.Info("开始获取当前用户的所有网络列表")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	networks, err := h.networkService.GetAllNetworks(userID.(string))
	if err != nil {
		logger.Error("获取网络列表失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功获取网络列表", zap.Int("network_count", len(networks)))

	return c.Status(fiber.StatusOK).JSON(networks)
}

// GetNetwork 获取特定网络
func (h *NetworkHandler) GetNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("开始获取特定网络", zap.String("network_id", id))

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	network, err := h.networkService.GetNetworkByID(id, userID.(string))
	if err != nil {
		logger.Error("获取网络失败", zap.String("network_id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if network == nil {
		logger.Warn("网络不存在或无权限访问", zap.String("network_id", id))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "网络不存在或无权限访问"})
	}

	logger.Info("成功获取网络", zap.String("network_id", id))

	return c.Status(fiber.StatusOK).JSON(network)
}

// CreateNetwork 创建新网络
func (h *NetworkHandler) CreateNetwork(c fiber.Ctx) error {
	var req zerotier.Network
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("创建网络请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始创建新网络", zap.String("network_name", req.Name))

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	network, err := h.networkService.CreateNetwork(&req, userID.(string))
	if err != nil {
		logger.Error("创建网络失败", zap.String("network_name", req.Name), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功创建网络", zap.String("network_id", network.ID), zap.String("network_name", network.Name))

	return c.Status(fiber.StatusCreated).JSON(network)
}

// UpdateNetwork 更新网络
func (h *NetworkHandler) UpdateNetwork(c fiber.Ctx) error {
	id := c.Params("id")

	var req zerotier.Network
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("更新网络请求参数绑定失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("开始更新网络", zap.String("network_id", id))

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	network, err := h.networkService.UpdateNetwork(id, &req, userID.(string))
	if err != nil {
		logger.Error("更新网络失败", zap.String("network_id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功更新网络", zap.String("network_id", network.ID))

	return c.Status(fiber.StatusOK).JSON(network)
}

// DeleteNetwork 删除网络
func (h *NetworkHandler) DeleteNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("开始删除网络", zap.String("network_id", id))

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	err := h.networkService.DeleteNetwork(id, userID.(string))
	if err != nil {
		logger.Error("删除网络失败", zap.String("network_id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功删除网络", zap.String("network_id", id))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "网络删除成功"})
}

// GetImportableNetworks 获取可导入的网络列表
func (h *NetworkHandler) GetImportableNetworks(c fiber.Ctx) error {
	logger.Info("开始获取可导入的网络列表")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	importableNetworks, err := h.networkService.GetImportableNetworks(userID.(string))
	if err != nil {
		logger.Error("获取可导入网络列表失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功获取可导入网络列表", zap.Int("count", len(importableNetworks)))

	return c.Status(fiber.StatusOK).JSON(importableNetworks)
}

// ImportNetworks 导入指定的网络
func (h *NetworkHandler) ImportNetworks(c fiber.Ctx) error {
	logger.Info("开始导入网络")

	// Get user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		logger.Error("获取用户ID失败")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权访问"})
	}

	// Parse request body
	var request struct {
		NetworkIDs []string `json:"network_ids"`
	}

	if err := c.Bind().Body(&request); err != nil {
		logger.Error("解析导入请求失败", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if len(request.NetworkIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "网络ID列表为空"})
	}

	logger.Info("开始导入网络", zap.Strings("network_ids", request.NetworkIDs))

	importedIDs, err := h.networkService.ImportNetworks(request.NetworkIDs, userID.(string))
	if err != nil {
		logger.Error("导入网络失败", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Info("成功导入网络", zap.Int("imported_count", len(importedIDs)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "导入成功",
		"imported_ids": importedIDs,
	})
}
