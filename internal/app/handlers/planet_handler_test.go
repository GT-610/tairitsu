package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GT-610/tairitsu/internal/app/config"
	"github.com/gofiber/fiber/v3"
)

const testIdentityPublic = "f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715"

func newPlanetHandlerForDir(dir string) *PlanetHandler {
	return NewPlanetHandler(&config.Config{ZeroTier: config.ZeroTierConfig{
		TokenPath: filepath.Join(dir, "authtoken.secret"),
	}})
}

func newSymlinkEscapeHandler(t *testing.T) (*PlanetHandler, string, string) {
	t.Helper()

	configuredDir := t.TempDir()
	outsideDir := t.TempDir()
	linkPath := filepath.Join(configuredDir, "outside-link")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}
	return newPlanetHandlerForDir(configuredDir), linkPath, outsideDir
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve path %q: %v", path, err)
	}
	return resolved
}

type generatePlanetWireResponse struct {
	Message               string `json:"message"`
	PlanetDataBase64      string `json:"planet_data"`
	PlanetID              uint64 `json:"planet_id"`
	BirthTime             int64  `json:"birth_time"`
	DownloadName          string `json:"download_name"`
	RootNodeCount         int    `json:"root_node_count"`
	EndpointCount         int    `json:"endpoint_count"`
	UsedRecommendedValues bool   `json:"used_recommended_values"`
}

func assertBase64PlanetPayload(t *testing.T, encoded string) {
	t.Helper()

	planetData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode planet_data Base64: %v", err)
	}
	if len(planetData) == 0 {
		t.Fatal("decoded planet_data is empty")
	}
}

func TestGetIdentityHandler_ReadsIdentityPublic(t *testing.T) {
	tempDir := t.TempDir()
	handler := newPlanetHandlerForDir(tempDir)

	identityPath := filepath.Join(tempDir, "identity.public")
	if err := os.WriteFile(identityPath, []byte(testIdentityPublic+"\n"), 0644); err != nil {
		t.Fatalf("write identity.public: %v", err)
	}

	app := fiber.New()
	app.Get("/identity", handler.GetIdentity)

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
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
	wantIdentityPath := filepath.Join(resolvedPath(t, tempDir), "identity.public")
	if body.IdentityPath != wantIdentityPath {
		t.Fatalf("identity_path = %q, want %q", body.IdentityPath, wantIdentityPath)
	}
}

func TestGetIdentityHandler_RejectsDirectoryOutsideConfiguredZeroTierPath(t *testing.T) {
	configuredDir := t.TempDir()
	handler := newPlanetHandlerForDir(configuredDir)

	app := fiber.New()
	app.Get("/identity", handler.GetIdentity)

	outsideDir := t.TempDir()
	req := httptest.NewRequest(http.MethodGet, "/identity?path="+url.QueryEscape(outsideDir), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestGetIdentityHandler_ReturnsNotFoundForMissingIdentity(t *testing.T) {
	tempDir := t.TempDir()
	handler := newPlanetHandlerForDir(tempDir)

	app := fiber.New()
	app.Get("/identity", handler.GetIdentity)

	req := httptest.NewRequest(http.MethodGet, "/identity?path="+tempDir, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
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

func TestGeneratePlanetHandler_ReturnsBase64PlanetBinaryAndMetadata(t *testing.T) {
	handler := NewPlanetHandler(nil)
	app := fiber.New()
	app.Post("/planet", handler.GeneratePlanet)

	body := `{"root_nodes":[{"identity_public":"f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715","endpoints":["203.0.113.1/9993"],"comments":"test"}],"recommend_values":true,"download_name":"planet.custom"}`
	req := httptest.NewRequest(http.MethodPost, "/planet", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var result generatePlanetWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.PlanetID == 0 {
		t.Fatalf("planet_id = 0, want non-zero")
	}
	if result.BirthTime <= 0 {
		t.Fatalf("birth_time = %d, want positive value", result.BirthTime)
	}
	assertBase64PlanetPayload(t, result.PlanetDataBase64)
	if result.DownloadName != "planet.custom" {
		t.Fatalf("download_name = %q, want planet.custom", result.DownloadName)
	}
	if result.RootNodeCount != 1 {
		t.Fatalf("root_node_count = %d, want 1", result.RootNodeCount)
	}
	if result.EndpointCount != 1 {
		t.Fatalf("endpoint_count = %d, want 1", result.EndpointCount)
	}
	if !result.UsedRecommendedValues {
		t.Fatal("used_recommended_values = false, want true")
	}
}

func TestGeneratePlanetHandler_RejectsDuplicateRootIdentity(t *testing.T) {
	handler := NewPlanetHandler(nil)
	app := fiber.New()
	app.Post("/planet", handler.GeneratePlanet)

	body := `{"root_nodes":[{"identity_public":"f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715","endpoints":["203.0.113.1/9993"]},{"identity_public":"f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715","endpoints":["203.0.113.2/9993"]}],"recommend_values":true}`
	req := httptest.NewRequest(http.MethodPost, "/planet", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestGetSigningKeysInfoHandler_ReturnsStatus(t *testing.T) {
	tempDir := t.TempDir()
	handler := newPlanetHandlerForDir(tempDir)

	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")
	if err := os.WriteFile(prevPath, []byte("ready"), 0644); err != nil {
		t.Fatalf("write previous key: %v", err)
	}
	if err := os.WriteFile(curPath, []byte("ready"), 0644); err != nil {
		t.Fatalf("write current key: %v", err)
	}

	app := fiber.New()
	app.Get("/signing-keys", handler.GetSigningKeysInfo)

	req := httptest.NewRequest(http.MethodGet, "/signing-keys?path="+tempDir, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
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
	wantSigningKeyPath := resolvedPath(t, tempDir)
	if body.SigningKeyPath != wantSigningKeyPath {
		t.Fatalf("signing_key_path = %q, want %q", body.SigningKeyPath, wantSigningKeyPath)
	}
}

func TestPlanetFilesystemHandlers_RejectSymlinkEscapes(t *testing.T) {
	testCases := []struct {
		name             string
		register         func(*fiber.App, *PlanetHandler)
		request          func(*testing.T, string) *http.Request
		checkNoKeysWrite bool
	}{
		{
			name: "identity",
			register: func(app *fiber.App, handler *PlanetHandler) {
				app.Get("/identity", handler.GetIdentity)
			},
			request: func(_ *testing.T, linkPath string) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/identity?path="+url.QueryEscape(linkPath), nil)
			},
		},
		{
			name: "signing key status",
			register: func(app *fiber.App, handler *PlanetHandler) {
				app.Get("/signing-keys", handler.GetSigningKeysInfo)
			},
			request: func(_ *testing.T, linkPath string) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/signing-keys?path="+url.QueryEscape(linkPath), nil)
			},
		},
		{
			name: "signing key generation",
			register: func(app *fiber.App, handler *PlanetHandler) {
				app.Post("/signing-keys", handler.GenerateSigningKeys)
			},
			request: func(_ *testing.T, linkPath string) *http.Request {
				return httptest.NewRequest(http.MethodPost, "/signing-keys?path="+url.QueryEscape(linkPath), nil)
			},
			checkNoKeysWrite: true,
		},
		{
			name: "custom signing key path",
			register: func(app *fiber.App, handler *PlanetHandler) {
				app.Post("/planet", handler.GeneratePlanet)
			},
			request: func(t *testing.T, linkPath string) *http.Request {
				t.Helper()
				payload, err := json.Marshal(map[string]any{
					"root_nodes": []map[string]any{{
						"identity_public": testIdentityPublic,
						"endpoints":       []string{"203.0.113.1/9993"},
					}},
					"signing_key_path": linkPath,
					"recommend_values": true,
				})
				if err != nil {
					t.Fatalf("marshal request: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/planet", bytes.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, linkPath, outsideDir := newSymlinkEscapeHandler(t)
			app := fiber.New()
			tc.register(app, handler)

			resp, err := app.Test(tc.request(t, linkPath))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
			}

			if tc.checkNoKeysWrite {
				for _, name := range []string{"previous.c25519", "current.c25519"} {
					if _, err := os.Stat(filepath.Join(outsideDir, name)); !os.IsNotExist(err) {
						t.Fatalf("%s was written through escaped symlink", name)
					}
				}
			}
		})
	}
}
