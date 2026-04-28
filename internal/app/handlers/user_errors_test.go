package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/gofiber/fiber/v3"
)

func TestWriteUserServiceError_MapsKnownErrors(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{name: "db unavailable", err: services.ErrUserDBUnavailable, expectedCode: fiber.StatusServiceUnavailable},
		{name: "username exists", err: services.ErrUsernameExists, expectedCode: fiber.StatusBadRequest},
		{name: "invalid credentials", err: services.ErrInvalidCredentials, expectedCode: fiber.StatusUnauthorized},
		{name: "user not found", err: services.ErrUserNotFound, expectedCode: fiber.StatusNotFound},
		{name: "old password incorrect", err: services.ErrOldPasswordIncorrect, expectedCode: fiber.StatusBadRequest},
		{name: "invalid user role", err: services.ErrInvalidUserRole, expectedCode: fiber.StatusBadRequest},
		{name: "transfer to self", err: services.ErrAdminTransferSelf, expectedCode: fiber.StatusBadRequest},
		{name: "delete self", err: services.ErrAdminDeleteSelf, expectedCode: fiber.StatusBadRequest},
		{name: "delete current admin blocked", err: services.ErrAdminDeleteBlocked, expectedCode: fiber.StatusBadRequest},
		{name: "target already admin", err: services.ErrTransferTargetAdmin, expectedCode: fiber.StatusBadRequest},
		{name: "admin access denied", err: services.ErrAdminAccessDenied, expectedCode: fiber.StatusForbidden},
		{name: "wrapped user not found", err: fmt.Errorf("wrapped: %w", services.ErrUserNotFound), expectedCode: fiber.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return writeUserServiceError(c, tc.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			if resp.StatusCode != tc.expectedCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.expectedCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			messageText, ok := body["message"].(string)
			if !ok || strings.TrimSpace(messageText) == "" {
				t.Fatalf("expected non-empty message body")
			}
			errorCode, ok := body["error_code"].(string)
			if !ok || strings.TrimSpace(errorCode) == "" {
				t.Fatalf("expected non-empty error_code body")
			}
		})
	}
}
