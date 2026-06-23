package handlers

import (
	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// NetworkHandler handles network-related HTTP requests
type NetworkHandler struct {
	networkService *services.NetworkService
}

// NewNetworkHandler creates a new network handler instance
func NewNetworkHandler(networkService *services.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// GetStatus retrieves the ZeroTier network status
func (h *NetworkHandler) GetStatus(c fiber.Ctx) error {
	logger.Info("Getting ZeroTier network status")

	status := h.networkService.GetRuntimeStatus()

	logger.Info("ZeroTier network status retrieved")

	return c.Status(fiber.StatusOK).JSON(status)
}

// GetNetworks retrieves all networks owned by the current user
func (h *NetworkHandler) GetNetworks(c fiber.Ctx) error {
	logger.Info("Getting networks for current user")

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	networks, err := h.networkService.GetAllNetworks(userID)
	if err != nil {
		logger.Error("Failed to get network list", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("Network list retrieved", zap.Int("network_count", len(networks)))

	return c.Status(fiber.StatusOK).JSON(networks)
}

func (h *NetworkHandler) GetSharedNetworks(c fiber.Ctx) error {
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	networks, err := h.networkService.GetSharedNetworks(userID)
	if err != nil {
		logger.Error("Failed to get shared network list", zap.String("user_id", userID), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(networks)
}

// GetNetwork retrieves a specific network
func (h *NetworkHandler) GetNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	if err := validateNetworkID(id); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	logger.Info("Getting network", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	network, err := h.networkService.GetNetworkByID(id, userID)
	if err != nil {
		logger.Error("Failed to get network", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network access denied")
	}

	logger.Info("Network retrieved", zap.String("network_id", id))

	return c.Status(fiber.StatusOK).JSON(network)
}

// CreateNetwork creates a new network
func (h *NetworkHandler) CreateNetwork(c fiber.Ctx) error {
	var req zerotier.Network
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind create network request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := validateNetworkName(req.Name); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateNetworkDescription(req.Description); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Creating network", zap.String("network_name", req.Name))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	network, err := h.networkService.CreateNetwork(&req, userID)
	if err != nil {
		logger.Error("Failed to create network", zap.String("network_name", req.Name), zap.Error(err))
		return writeErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	logger.Info("Network created", zap.String("network_id", network.ID), zap.String("network_name", network.Name))

	return c.Status(fiber.StatusCreated).JSON(network)
}

// UpdateNetwork updates a network
func (h *NetworkHandler) UpdateNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	if err := validateNetworkID(id); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	var req zerotier.NetworkUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind update network request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Updating network", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	network, err := h.networkService.UpdateNetwork(id, &req, userID)
	if err != nil {
		logger.Error("Failed to update network", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network update access denied")
	}

	logger.Info("Network updated", zap.String("network_id", network.ID))

	return c.Status(fiber.StatusOK).JSON(network)
}

// UpdateNetworkMetadata updates network metadata (name and description)
// Name is synced to both the controller and the database; description is database-only
func (h *NetworkHandler) UpdateNetworkMetadata(c fiber.Ctx) error {
	id := c.Params("id")
	if err := validateNetworkID(id); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind().Body(&req); err != nil {
		logger.Error("Failed to bind network metadata update request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := validateNetworkName(req.Name); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateNetworkDescription(req.Description); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Updating network metadata", zap.String("network_id", id), zap.String("name", req.Name))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	network, err := h.networkService.UpdateNetworkMetadata(id, req.Name, req.Description, userID)
	if err != nil {
		logger.Error("Failed to update network metadata", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network update access denied")
	}

	logger.Info("Network metadata updated", zap.String("network_id", network.ID))

	return c.Status(fiber.StatusOK).JSON(network)
}

// DeleteNetwork deletes a network
func (h *NetworkHandler) DeleteNetwork(c fiber.Ctx) error {
	id := c.Params("id")
	if err := validateNetworkID(id); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	logger.Info("Deleting network", zap.String("network_id", id))

	// Get user ID from context
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	err := h.networkService.DeleteNetwork(id, userID)
	if err != nil {
		logger.Error("Failed to delete network", zap.String("network_id", id), zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network delete access denied")
	}

	logger.Info("Network deleted", zap.String("network_id", id))

	return writeMessageResponse(c, fiber.StatusOK, "network.delete_success", "Network deleted successfully", nil)
}

// GetImportableNetworks retrieves the list of importable networks
func (h *NetworkHandler) GetImportableNetworks(c fiber.Ctx) error {
	logger.Info("Getting importable networks")

	// Get user ID from context
	_, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	importableNetworks, err := h.networkService.GetImportableNetworks()
	if err != nil {
		logger.Error("Failed to get importable networks", zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Importable network access denied")
	}

	logger.Info("Importable networks retrieved", zap.Int("count", len(importableNetworks.Candidates)))
	return c.Status(fiber.StatusOK).JSON(importableNetworks)
}

// ImportNetworks imports the specified networks
func (h *NetworkHandler) ImportNetworks(c fiber.Ctx) error {
	logger.Info("Importing networks")

	// Get user ID from context
	_, authErr := requiredUserID(c)
	if authErr != nil {
		logger.Error("Failed to get user ID")
		return authErr
	}

	// Parse request body
	var request struct {
		NetworkIDs []string `json:"network_ids"`
		OwnerID    string   `json:"owner_id"`
	}

	if err := c.Bind().Body(&request); err != nil {
		logger.Error("Failed to bind import networks request", zap.Error(err))
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if len(request.NetworkIDs) == 0 {
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "network.import_empty", "Network ID list is empty")
	}

	role, _ := c.Locals("role").(string)

	logger.Info("Processing network import", zap.Strings("network_ids", request.NetworkIDs), zap.String("owner_id", request.OwnerID))

	result, err := h.networkService.ImportNetworks(request.NetworkIDs, request.OwnerID, role)
	if err != nil {
		logger.Error("Network import failed", zap.Error(err))
		return writeNetworkServiceError(c, err, "Network not found", "Network import access denied")
	}

	logger.Info("Network import completed",
		zap.Int("imported_count", len(result.Imported)),
		zap.Int("skipped_count", len(result.Skipped)),
		zap.Int("failed_count", len(result.Failed)))

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *NetworkHandler) GetNetworkViewers(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	viewers, err := h.networkService.GetNetworkViewers(networkID, userID)
	if err != nil {
		return writeNetworkServiceError(c, err, "Network not found", "Network viewer access denied")
	}

	return c.Status(fiber.StatusOK).JSON(viewers)
}

func (h *NetworkHandler) GetNetworkViewerCandidates(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}

	candidates, err := h.networkService.GetNetworkViewerCandidates(networkID, userID)
	if err != nil {
		return writeNetworkServiceError(c, err, "Network not found", "Network viewer access denied")
	}

	return c.Status(fiber.StatusOK).JSON(candidates)
}

func (h *NetworkHandler) AddNetworkViewer(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
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
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.required", "User is required")
	}

	if err := h.networkService.GrantNetworkViewer(networkID, request.UserID, userID); err != nil {
		return writeNetworkServiceError(c, err, "Network not found", "Network viewer access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "network.viewer_added", "Read-only viewer access granted", nil)
}

func (h *NetworkHandler) DeleteNetworkViewer(c fiber.Ctx) error {
	networkID := c.Params("id")
	if err := validateNetworkID(networkID); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	targetUserID := c.Params("userId")
	userID, authErr := requiredUserID(c)
	if authErr != nil {
		return authErr
	}
	if targetUserID == "" {
		return writeErrorResponseWithCode(c, fiber.StatusBadRequest, "user.required", "User is required")
	}

	if err := h.networkService.RevokeNetworkViewer(networkID, targetUserID, userID); err != nil {
		return writeNetworkServiceError(c, err, "Network not found", "Network viewer access denied")
	}

	return writeMessageResponse(c, fiber.StatusOK, "network.viewer_removed", "Read-only viewer access removed", nil)
}
