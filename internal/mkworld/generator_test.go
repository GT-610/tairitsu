package mkworld

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	validIdentityPublic       = "f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715"
	secondValidIdentityPublic = "6a4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9726"
)

func testRootNode(identity string, endpoints ...string) RootNodeConfig {
	return RootNodeConfig{
		IdentityPublic: identity,
		Comments:       "test root",
		Endpoints:      endpoints,
	}
}

func TestGeneratePlanet_RejectsInvalidInput(t *testing.T) {
	testCases := []struct {
		name    string
		options *GenerateOptions
		target  error
	}{
		{
			name:    "missing root nodes",
			options: &GenerateOptions{},
			target:  ErrNoRootNodes,
		},
		{
			name: "missing identity",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{{Endpoints: []string{"203.0.113.1/9993"}}},
				RecommendValues: true,
			},
			target: ErrIdentityPublicRequired,
		},
		{
			name: "missing endpoints",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{{IdentityPublic: validIdentityPublic}},
				RecommendValues: true,
			},
			target: ErrNoEndpoints,
		},
		{
			name: "invalid identity",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode("invalid", "203.0.113.1/9993")},
				RecommendValues: true,
			},
			target: ErrInvalidIdentity,
		},
		{
			name: "invalid endpoint",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "not-an-endpoint")},
				RecommendValues: true,
			},
			target: ErrInvalidEndpoint,
		},
		{
			name: "duplicate endpoint",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993", "203.0.113.1/9993")},
				RecommendValues: true,
			},
			target: ErrDuplicateEndpoint,
		},
		{
			name: "duplicate identity",
			options: &GenerateOptions{
				RootNodes: []RootNodeConfig{
					testRootNode(validIdentityPublic, "203.0.113.1/9993"),
					testRootNode(validIdentityPublic, "203.0.113.2/9993"),
				},
				RecommendValues: true,
			},
			target: ErrDuplicateIdentity,
		},
		{
			name: "reserved planet id",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
				PlanetID:        uint64(ZT_WORLD_ID_EARTH),
				BirthTime:       time.Now().UnixMilli(),
				RecommendValues: false,
			},
			target: ErrReservedPlanetID,
		},
		{
			name: "invalid birth time",
			options: &GenerateOptions{
				RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
				PlanetID:        123456789,
				BirthTime:       1567191349589,
				RecommendValues: false,
			},
			target: ErrInvalidBirthTime,
		},
		{
			name: "too many root nodes",
			options: &GenerateOptions{
				RootNodes: []RootNodeConfig{
					testRootNode(validIdentityPublic, "203.0.113.1/9993"),
					testRootNode(secondValidIdentityPublic, "203.0.113.2/9993"),
					testRootNode("7b4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9727", "203.0.113.3/9993"),
					testRootNode("8c4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9728", "203.0.113.4/9993"),
					testRootNode("9d4d8f1c22:0:4c2b7d5e9f4a3c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9729", "203.0.113.5/9993"),
				},
				RecommendValues: true,
			},
			target: ErrMaxRootNodesExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GeneratePlanet(tc.options)
			if !errors.Is(err, tc.target) {
				t.Fatalf("GeneratePlanet() error = %v, want %v", err, tc.target)
			}
		})
	}
}

func TestGeneratePlanet_ReturnsMetadataAndData(t *testing.T) {
	result, err := GeneratePlanet(&GenerateOptions{
		RootNodes: []RootNodeConfig{
			testRootNode(validIdentityPublic, "203.0.113.1/9993", "2001:db8::1/9993"),
			testRootNode(secondValidIdentityPublic, "203.0.113.2/9993"),
		},
		RecommendValues: true,
		DownloadName:    "planet.custom",
	})
	if err != nil {
		t.Fatalf("GeneratePlanet() error = %v", err)
	}

	if result.PlanetID == 0 {
		t.Fatalf("PlanetID = 0, want non-zero")
	}
	if result.BirthTime <= 0 {
		t.Fatalf("BirthTime = %d, want positive value", result.BirthTime)
	}
	if len(result.PlanetData) == 0 {
		t.Fatalf("PlanetData is empty")
	}
	if result.RootNodeCount != 2 {
		t.Fatalf("RootNodeCount = %d, want 2", result.RootNodeCount)
	}
	if result.EndpointCount != 3 {
		t.Fatalf("EndpointCount = %d, want 3", result.EndpointCount)
	}
	if result.DownloadName != "planet.custom" {
		t.Fatalf("DownloadName = %q, want planet.custom", result.DownloadName)
	}
	if !result.UsedRecommendedValues {
		t.Fatal("UsedRecommendedValues = false, want true")
	}
}

func TestGeneratePlanet_UsesCustomMetadataAndSigningKeys(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	if err := CreateSigningKeys(prevPath, curPath); err != nil {
		t.Fatalf("CreateSigningKeys() error = %v", err)
	}

	result, err := GeneratePlanet(&GenerateOptions{
		RootNodes:       []RootNodeConfig{testRootNode(validIdentityPublic, "203.0.113.1/9993")},
		SigningKeyPath:  tempDir,
		PlanetID:        123456789,
		BirthTime:       time.Now().UnixMilli(),
		RecommendValues: false,
	})
	if err != nil {
		t.Fatalf("GeneratePlanet() error = %v", err)
	}

	if result.UsedRecommendedValues {
		t.Fatal("UsedRecommendedValues = true, want false")
	}
}

func TestReadSigningKeys_RejectsInvalidLength(t *testing.T) {
	tempDir := t.TempDir()
	prevPath := filepath.Join(tempDir, "previous.c25519")
	curPath := filepath.Join(tempDir, "current.c25519")

	if err := os.WriteFile(prevPath, []byte("bad"), 0644); err != nil {
		t.Fatalf("write previous key: %v", err)
	}
	if err := os.WriteFile(curPath, []byte("bad"), 0644); err != nil {
		t.Fatalf("write current key: %v", err)
	}

	if _, _, _, _, err := ReadSigningKeys(prevPath, curPath); err == nil {
		t.Fatal("ReadSigningKeys() error = nil, want error")
	}
}
