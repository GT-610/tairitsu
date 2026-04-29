package middleware

import "github.com/gofiber/fiber/v3"

type initializationState interface {
	IsInitialized() bool
}

// SetupOnlyWithState blocks requests once the application has completed initialization.
func SetupOnlyWithState(state initializationState) fiber.Handler {
	return func(c fiber.Ctx) error {
		if state.IsInitialized() {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:     "Already Initialized",
				Message:   "The system is already initialized. This endpoint is only available during first-time setup.",
				ErrorCode: "system.already_initialized",
				Code:      fiber.StatusConflict,
			})
		}

		return c.Next()
	}
}

// InitializedOnlyWithState blocks runtime routes until the application has completed setup.
func InitializedOnlyWithState(state initializationState) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !state.IsInitialized() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Error:     "Setup Required",
				Message:   "System setup is required. Complete the setup wizard first.",
				ErrorCode: "system.setup_required",
				Code:      fiber.StatusServiceUnavailable,
			})
		}

		return c.Next()
	}
}
