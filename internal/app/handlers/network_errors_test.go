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

func TestWriteNetworkServiceError_MapsWrappedImportErrors(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "wrapped import access denied",
			err:          fmt.Errorf("wrapped: %w", services.ErrImportAccessDenied),
			expectedCode: fiber.StatusForbidden,
		},
		{
			name:         "wrapped import owner required",
			err:          fmt.Errorf("wrapped: %w", services.ErrImportOwnerRequired),
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name:         "wrapped import owner not found",
			err:          fmt.Errorf("wrapped: %w", services.ErrImportOwnerNotFound),
			expectedCode: fiber.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return writeNetworkServiceError(c, tc.err, "网络不存在", "无权限访问网络")
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
