package mkworld

import (
	"errors"
	"testing"
)

const validIdentityPublic = "f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715"

func TestGeneratePlanet_RejectsInvalidInput(t *testing.T) {
	testCases := []struct {
		name    string
		options *GenerateOptions
		target  error
	}{
		{
			name: "missing identity",
			options: &GenerateOptions{
				IdentityPublic: "",
				Endpoints:      []string{"203.0.113.1/9993"},
			},
			target: ErrIdentityPublicRequired,
		},
		{
			name: "missing endpoints",
			options: &GenerateOptions{
				IdentityPublic: validIdentityPublic,
				Endpoints:      nil,
			},
			target: ErrNoEndpoints,
		},
		{
			name: "invalid identity",
			options: &GenerateOptions{
				IdentityPublic: "invalid",
				Endpoints:      []string{"203.0.113.1/9993"},
			},
			target: ErrInvalidIdentity,
		},
		{
			name: "invalid endpoint",
			options: &GenerateOptions{
				IdentityPublic: validIdentityPublic,
				Endpoints:      []string{"not-an-endpoint"},
			},
			target: ErrInvalidEndpoint,
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
		IdentityPublic: validIdentityPublic,
		Endpoints:      []string{"203.0.113.1/9993", "2001:db8::1/9993"},
		Comments:       "test planet",
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
}
