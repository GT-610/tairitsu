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
	if err := os.WriteFile(identityPath, []byte("f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715\n"), 0644); err != nil {
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

	body := `{"identity_public":"f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715","endpoints":["203.0.113.1/9993"],"comments":"test"}`
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

func TestGetSigningKeysInfoHandler_ReturnsStatus(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")
	if err := os.WriteFile(prevPath, []byte("ready"), 0644); err != nil {
		t.Fatalf("write previous key: %v", err)
	}
	if err := os.WriteFile(curPath, []byte("ready"), 0644); err != nil {
		t.Fatalf("write current key: %v", err)
	}

	app := fiber.New()
	app.Get("/signing-keys", GetSigningKeysInfoHandler)

	req := httptest.NewRequest(http.MethodGet, "/signing-keys?path="+tempDir, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body SigningKeysInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Ready {
		t.Fatal("Ready = false, want true")
	}
	if body.SigningKeyPath != tempDir {
		t.Fatalf("signing_key_path = %q, want %q", body.SigningKeyPath, tempDir)
	}
}
