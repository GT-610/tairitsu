package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getWithEncoding requests an endpoint with the given Accept-Encoding. An empty
// encoding means the header is omitted entirely.
func getWithEncoding(t *testing.T, app *fiber.App, path, encoding string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if encoding != "" {
		req.Header.Set(fiber.HeaderAcceptEncoding, encoding)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	return resp
}

// The console bundle shares the process with the API, so the JSON surface must
// still be compressed. /api/system/status is the largest unauthenticated
// response in the real route table.
func TestRoutesCompressJSONWithGzip(t *testing.T) {
	app, _ := newTestApp(t, initialized(true))

	resp := getWithEncoding(t, app, "/api/system/status", "gzip")

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, strings.ToLower(resp.Header.Get(fiber.HeaderContentEncoding)), "gzip")
	assert.Contains(t, resp.Header.Get(fiber.HeaderVary), fiber.HeaderAcceptEncoding)
}

// Brotli is negotiated ahead of gzip when the client advertises it, which is
// what browsers do and what the console bundle benefits from most.
func TestRoutesCompressJSONWithBrotli(t *testing.T) {
	app, _ := newTestApp(t, initialized(true))

	resp := getWithEncoding(t, app, "/api/system/status", "br")

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, strings.ToLower(resp.Header.Get(fiber.HeaderContentEncoding)), "br")
	assert.Contains(t, resp.Header.Get(fiber.HeaderVary), fiber.HeaderAcceptEncoding)
}

// The embedded app shell is served by the same middleware stack, so a page load
// is compressed too. A checkout without a frontend build serves the placeholder
// page instead, which is also compressed, so the only assertion that holds in
// both states is that the HTML body is encoded.
func TestRoutesCompressEmbeddedIndexHTML(t *testing.T) {
	app, _ := newTestApp(t, initialized(true))

	resp := getWithEncoding(t, app, "/", "br")

	assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), "text/html")
	assert.Contains(t, strings.ToLower(resp.Header.Get(fiber.HeaderContentEncoding)), "br")

	// A real build serves the app shell, which must stay uncached so a new
	// release is picked up on the next load. The placeholder is a 503 and
	// carries no cache policy.
	if resp.StatusCode == fiber.StatusOK {
		assert.Equal(t, "no-cache", resp.Header.Get(fiber.HeaderCacheControl))
	}
}

// A client that cannot decode compressed bodies must still receive a readable,
// uncompressed response.
func TestRoutesSkipCompressionWithoutAcceptEncoding(t *testing.T) {
	app, _ := newTestApp(t, initialized(true))

	resp := getWithEncoding(t, app, "/api/system/status", "")

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get(fiber.HeaderContentEncoding))
	assert.Contains(t, resp.Header.Get(fiber.HeaderVary), fiber.HeaderAcceptEncoding)
}

// Bodies below the compressor's threshold are not worth encoding, so the
// liveness probe stays unencoded. This also guards the case that matters for
// health checks: a probe must never depend on decompression support.
func TestRoutesSkipCompressionForTinyBody(t *testing.T) {
	app, _ := newTestApp(t, initialized(true))

	resp := getWithEncoding(t, app, "/api/health", "gzip")

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get(fiber.HeaderContentEncoding))
}
