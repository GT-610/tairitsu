package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIPathMatchesTheJSONSurfaceOnly(t *testing.T) {
	for _, path := range []string{"/api", "/api/", "/api/networks", "/api/system/status"} {
		assert.True(t, isAPIPath(path), path)
	}

	for _, path := range []string{"/", "/apiary", "/networks", "/networks/8056c2e21c000001", "/apix/networks"} {
		assert.False(t, isAPIPath(path), path)
	}
}

// The distinction decides whether an unknown path gets the app shell, so a
// client route must never be mistaken for a file.
func TestNamesAFileSeparatesRoutesFromAssetRequests(t *testing.T) {
	for _, path := range []string{
		"/vite.svg",
		"/favicon.ico",
		"/assets/index-CKMRbpY-.js",
		"/assets/index-BPkJeLui.css",
		"/index.html",
		"/.well-known/security.txt",
	} {
		assert.True(t, namesAFile(path), path)
	}

	for _, path := range []string{
		"/",
		"/setup",
		"/login",
		"/register",
		"/networks",
		"/networks/8056c2e21c000001",
		"/networks/shared/8056c2e21c000001",
		"/user-management",
		"/import-network",
		"/planet",
		"/settings",
		"/dashboard",
	} {
		assert.False(t, namesAFile(path), path)
	}
}
