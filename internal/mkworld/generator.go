/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PlanetConfig struct {
	Comments  string   `json:"comments"`
	Endpoints []string `json:"endpoints"`
}

type GenerateOptions struct {
	IdentityPublic string
	Endpoints      []string
	Comments       string
}

type GeneratedPlanet struct {
	PlanetID   uint64
	BirthTime  int64
	PlanetData []byte
}

var (
	ErrIdentityPublicRequired = errors.New("identity.public is required")
	ErrNoEndpoints            = errors.New("at least one endpoint is required")
)

func GeneratePlanet(opts *GenerateOptions) (*GeneratedPlanet, error) {
	if opts.IdentityPublic == "" {
		return nil, ErrIdentityPublicRequired
	}

	identity, err := ParseIdentityPublic(opts.IdentityPublic)
	if err != nil {
		return nil, err
	}

	endpointValues := normalizeEndpoints(opts.Endpoints)
	if len(endpointValues) == 0 {
		return nil, ErrNoEndpoints
	}

	endpoints := make([]*ZtNodeInetAddr, 0, len(endpointValues))
	for _, epStr := range endpointValues {
		ep := &ZtNodeInetAddr{}
		if err := ep.FromString(epStr); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEndpoint, epStr)
		}
		endpoints = append(endpoints, ep)
	}

	prevPub, prevPriv := GenerateSigningKeyPair()
	curPub := prevPub

	planetID, err := generatePlanetID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate planet id: %w", err)
	}
	planetBirth := time.Now().UnixMilli()

	ztW := &ZtWorld{
		Type:      ZT_WORLD_TYPE_PLANET,
		ID:        ZtWorldID(planetID),
		Timestamp: uint64(planetBirth),
	}

	rootNode := &ZtWorldPlanetNode{
		Identity:  identity,
		Endpoints: endpoints,
	}
	ztW.Nodes = []*ZtWorldPlanetNode{rootNode}

	ztW.PublicKeyMustBeSignedByNextTime = curPub

	toSignData, err := ztW.Serialize(true, [64]byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize for signing: %w", err)
	}

	sig, err := SignMessage(prevPub, prevPriv, toSignData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	finalData, err := ztW.Serialize(false, sig)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize final: %w", err)
	}

	return &GeneratedPlanet{
		PlanetID:   planetID,
		BirthTime:  planetBirth,
		PlanetData: finalData,
	}, nil
}

func ParseIdentityPublic(s string) (*ZtWorldPlanetNodeIdentity, error) {
	parts := splitByColon(s)
	if len(parts) < 3 {
		return nil, fmt.Errorf("%w: expected format address:0:publicKey", ErrInvalidIdentity)
	}

	identity := &ZtWorldPlanetNodeIdentity{}

	addrBytes, err := hex.DecodeString(parts[0])
	if err != nil || len(addrBytes) != 5 {
		return nil, fmt.Errorf("%w: invalid address", ErrInvalidIdentity)
	}
	copy(identity.ZtNodeAddress[:], addrBytes)

	if parts[1] != "0" {
		return nil, fmt.Errorf("%w: invalid separator field", ErrInvalidIdentity)
	}

	pubBytes, err := hex.DecodeString(parts[2])
	if err != nil || len(pubBytes) != 32 {
		return nil, fmt.Errorf("%w: invalid public key", ErrInvalidIdentity)
	}
	copy(identity.PublicKey[:], pubBytes)

	return identity, nil
}

func splitByColon(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func normalizeEndpoints(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func generatePlanetID() (uint64, error) {
	b := make([]byte, 4)
	for {
		if _, err := rand.Read(b); err != nil {
			return 0, err
		}
		value := uint64(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
		if value != 0 && value != uint64(ZT_WORLD_ID_EARTH) && value != uint64(ZT_WORLD_ID_MARS) {
			return value, nil
		}
	}
}

type PlanetFile struct {
	Path        string
	Content     []byte
	PlanetID    uint64
	PlanetBirth uint64
	NodeID      string
}

func ReadPlanetFile(path string) (*PlanetFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read planet file: %w", err)
	}
	return &PlanetFile{Path: path, Content: data}, nil
}

func SavePlanetFile(data []byte, path string) error {
	return os.WriteFile(path, data, 0644)
}

func CreateSigningKeys(prevPath, curPath string) error {
	prevPub, prevPriv := GenerateSigningKeyPair()

	if err := writeKeyFile(prevPath, prevPub, prevPriv); err != nil {
		return fmt.Errorf("failed to write previous key: %w", err)
	}
	if err := writeKeyFile(curPath, prevPub, prevPriv); err != nil {
		return fmt.Errorf("failed to write current key: %w", err)
	}

	return nil
}

func writeKeyFile(path string, pub [32]byte, priv [32]byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(pub[:]); err != nil {
		return err
	}
	if _, err := file.Write(priv[:]); err != nil {
		return err
	}

	return nil
}

func ReadSigningKeys(prevPath, curPath string) (prevPub, curPub [32]byte, prevPriv, curPriv [32]byte, err error) {
	prevData, err := os.ReadFile(prevPath)
	if err != nil {
		return
	}
	curData, err := os.ReadFile(curPath)
	if err != nil {
		return
	}

	if len(prevData) != 64 || len(curData) != 64 {
		err = fmt.Errorf("invalid key file length")
		return
	}

	copy(prevPub[:], prevData[:32])
	copy(prevPriv[:], prevData[32:64])
	copy(curPub[:], curData[:32])
	copy(curPriv[:], curData[32:64])

	return
}

func ReadIdentityPublic(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read identity.public: %w", err)
	}
	return string(data), nil
}

func EnsureDirectory(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func ReadIdentityFromReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read identity: %w", err)
	}
	return string(data), nil
}
