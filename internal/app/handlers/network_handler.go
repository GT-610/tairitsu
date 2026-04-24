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

	status := h.networkService.GetRuntimeStatus()

	logger.Info("成功获取ZeroTier网络状态")

	return c.Status(fiber.StatusOK).JSON(status)
}

// GetNetworks 获取当前用户的所有网络
func (h *NetworkHandler) GetNetworks(c fiber.Ctx) error {
	logger.Info("开始获取当前用户的所有网络列表")

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	networks, err := h.networkService.GetAllNetworks(userID)
	if err != nil {
		logger.Error("获取网络列表失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("成功获取网络列表", zap.Int("network_count", len(networks)))

	return c.Status(fiber.StatusOK).JSON(networks)
}

func (h *NetworkHandler) GetSharedNetworks(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	networks, err := h.networkService.GetSharedNetworks(userID)
	if err != nil {
		logger.Error("获取共享网络列表失败", zap.String("user_id", userID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(networks)
}

// GetNetwork 获取特定网络
func (h *NetworkHandler) GetNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("开始获取特定网络", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	network, err := h.networkService.GetNetworkByID(id, userID)
	if err != nil {
		logger.Error("获取网络失败", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限访问网络")
	}

	logger.Info("成功获取网络", zap.String("network_id", id))

	return c.Status(fiber.StatusOK).JSON(network)
}

// CreateNetwork 创建新网络
func (h *NetworkHandler) CreateNetwork(c fiber.Ctx) error {
	var req zerotier.Network
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("创建网络请求参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("开始创建新网络", zap.String("network_name", req.Name))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	network, err := h.networkService.CreateNetwork(&req, userID)
	if err != nil {
		logger.Error("创建网络失败", zap.String("network_name", req.Name), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("成功创建网络", zap.String("network_id", network.ID), zap.String("network_name", network.Name))

	return c.Status(fiber.StatusCreated).JSON(network)
}

// UpdateNetwork 更新网络
func (h *NetworkHandler) UpdateNetwork(c fiber.Ctx) error {
	id := c.Params("id")

	var req zerotier.NetworkUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("更新网络请求参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("开始更新网络", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	network, err := h.networkService.UpdateNetwork(id, &req, userID)
	if err != nil {
		logger.Error("更新网络失败", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限更新网络")
	}

	logger.Info("成功更新网络", zap.String("network_id", network.ID))

	return c.Status(fiber.StatusOK).JSON(network)
}

// UpdateNetworkMetadata 更新网络元数据（名称和描述）
// 名称同步更新控制器和数据库，描述仅更新数据库
func (h *NetworkHandler) UpdateNetworkMetadata(c fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("更新网络元数据请求参数绑定失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("开始更新网络元数据", zap.String("network_id", id), zap.String("name", req.Name))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	network, err := h.networkService.UpdateNetworkMetadata(id, req.Name, req.Description, userID)
	if err != nil {
		logger.Error("更新网络元数据失败", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限更新网络")
	}

	logger.Info("成功更新网络元数据", zap.String("network_id", network.ID))

	return c.Status(fiber.StatusOK).JSON(network)
}

// DeleteNetwork 删除网络
func (h *NetworkHandler) DeleteNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("开始删除网络", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	err := h.networkService.DeleteNetwork(id, userID)
	if err != nil {
		logger.Error("删除网络失败", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限删除网络")
	}

	logger.Info("成功删除网络", zap.String("network_id", id))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "网络删除成功"})
}

// GetImportableNetworks 获取可导入的网络列表
func (h *NetworkHandler) GetImportableNetworks(c fiber.Ctx) error {
	logger.Info("开始获取可导入的网络列表")

	// Get user ID from context
	_, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	importableNetworks, err := h.networkService.GetImportableNetworks()
	if err != nil {
		logger.Error("获取可导入网络列表失败", zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限获取可导入网络")
	}

	logger.Info("成功获取可导入网络列表", zap.Int("count", len(importableNetworks.Candidates)))
	return c.Status(fiber.StatusOK).JSON(importableNetworks)
}

// ImportNetworks 导入指定的网络
func (h *NetworkHandler) ImportNetworks(c fiber.Ctx) error {
	logger.Info("开始导入网络")

	// Get user ID from context
	_, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("获取用户ID失败")
		return authErr
	}

	// Parse request body
	var request struct {
		NetworkIDs []string `json:"network_ids"`
		OwnerID    string   `json:"owner_id"`
	}

	if err := c.Bind().Body(&request); err != nil {
		logger.Error("解析导入请求失败", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if len(request.NetworkIDs) == 0 {
		return writeErrorResponse(c, fiber.StatusBadRequest, "网络ID列表为空")
	}

	role, _ := c.Locals("role").(string)

	logger.Info("开始导入网络", zap.Strings("network_ids", request.NetworkIDs), zap.String("owner_id", request.OwnerID))

	result, err := h.networkService.ImportNetworks(request.NetworkIDs, request.OwnerID, role)
	if err != nil {
		logger.Error("导入网络失败", zap.Error(err))
		return writeNetworkServiceError(c, err, "网络不存在", "无权限导入网络")
	}

	logger.Info("导入网络处理完成",
		zap.Int("imported_count", len(result.Imported)),
		zap.Int("skipped_count", len(result.Skipped)),
		zap.Int("failed_count", len(result.Failed)))

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *NetworkHandler) GetNetworkViewers(c fiber.Ctx) error {
	networkID := c.Params("id")
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	viewers, err := h.networkService.GetNetworkViewers(networkID, userID)
	if err != nil {
		return writeNetworkServiceError(c, err, "网络不存在", "无权限管理网络查看授权")
	}

	return c.Status(fiber.StatusOK).JSON(viewers)
}

func (h *NetworkHandler) GetNetworkViewerCandidates(c fiber.Ctx) error {
	networkID := c.Params("id")
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	candidates, err := h.networkService.GetNetworkViewerCandidates(networkID, userID)
	if err != nil {
		return writeNetworkServiceError(c, err, "网络不存在", "无权限管理网络查看授权")
	}

	return c.Status(fiber.StatusOK).JSON(candidates)
}

func (h *NetworkHandler) AddNetworkViewer(c fiber.Ctx) error {
	networkID := c.Params("id")
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	var request struct {
		UserID string `json:"user_id"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if request.UserID == "" {
		return writeErrorResponse(c, fiber.StatusBadRequest, "必须指定用户")
	}

	if err := h.networkService.GrantNetworkViewer(networkID, request.UserID, userID); err != nil {
		return writeNetworkServiceError(c, err, "网络不存在", "无权限管理网络查看授权")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "已授予只读查看权限"})
}

func (h *NetworkHandler) DeleteNetworkViewer(c fiber.Ctx) error {
	networkID := c.Params("id")
	targetUserID := c.Params("userId")
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}
	if targetUserID == "" {
		return writeErrorResponse(c, fiber.StatusBadRequest, "必须指定用户")
	}

	if err := h.networkService.RevokeNetworkViewer(networkID, targetUserID, userID); err != nil {
		return writeNetworkServiceError(c, err, "网络不存在", "无权限管理网络查看授权")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "已移除只读查看权限"})
}
