package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestGetIdentityHandler_ReadsIdentityPublic(t *testing.T) {
	tempDir := t.TempDir()
	identityPath := filepath.Join(tempDir, "identity.public")
	if err := os.WriteFile(identityPath, []byte("992fcf1db7:0:206ed59350b31916f749a1f85dffb3a8787dcbf83b8c6e9448d4e3ea0e336930\n"), 0644); err != nil {
		t.Fatalf("write identity.public: %v", err)
	}

	app := fiber.New()
	app.Get("/identity", GetIdentityHandler)

	req := httptest.NewRequest(http.MethodGet, "/identity?path="+tempDir, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body IdentityInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.IdentityPublic == "" {
		t.Fatalf("expected identity_public in response")
	}
	if body.IdentityPath != identityPath {
		t.Fatalf("identity_path = %q, want %q", body.IdentityPath, identityPath)
	}
}

func TestGetIdentityHandler_ReturnsNotFoundForMissingIdentity(t *testing.T) {
	tempDir := t.TempDir()

	app := fiber.New()
	app.Get("/identity", GetIdentityHandler)

	req := httptest.NewRequest(http.MethodGet, "/identity?path="+tempDir, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error body")
	}
	if !strings.HasSuffix(body["identity_path"], "/identity.public") {
		t.Fatalf("identity_path = %q, want suffix /identity.public", body["identity_path"])
	}
}

func TestGeneratePlanetHandler_ReturnsPlanetDataAndMetadata(t *testing.T) {
	app := fiber.New()
	app.Post("/planet", GeneratePlanetHandler)

	body := `{"identity_public":"992fcf1db7:0:206ed59350b31916f749a1f85dffb3a8787dcbf83b8c6e9448d4e3ea0e336930","endpoints":["203.0.113.1/9993"],"comments":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/planet", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var result GeneratePlanetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.PlanetID == 0 {
		t.Fatalf("planet_id = 0, want non-zero")
	}
	if result.BirthTime <= 0 {
		t.Fatalf("birth_time = %d, want positive value", result.BirthTime)
	}
	if len(result.PlanetData) == 0 {
		t.Fatalf("planet_data is empty")
	}
	if result.DownloadName != "planet" {
		t.Fatalf("download_name = %q, want planet", result.DownloadName)
	}
}
