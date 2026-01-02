/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	OutputPath     string
}

func GeneratePlanet(opts *GenerateOptions) ([]byte, error) {
	if opts.IdentityPublic == "" {
		return nil, fmt.Errorf("identity.public is required")
	}

	identity, err := ParseIdentityPublic(opts.IdentityPublic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	endpoints := make([]*ZtNodeInetAddr, 0, len(opts.Endpoints))
	for _, epStr := range opts.Endpoints {
		ep := &ZtNodeInetAddr{}
		if err := ep.FromString(epStr); err != nil {
			return nil, fmt.Errorf("failed to parse endpoint %s: %w", epStr, err)
		}
		endpoints = append(endpoints, ep)
	}

	prevPub, prevPriv := GenerateSigningKeyPair()
	curPub := prevPub

	planetID := generatePlanetID()
	planetBirth := uint64(time.Now().UnixMilli())

	ztW := &ZtWorld{
		Type:      ZT_WORLD_TYPE_PLANET,
		ID:        ZtWorldID(planetID),
		Timestamp: planetBirth,
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

	if opts.OutputPath != "" {
		if err := os.WriteFile(opts.OutputPath, finalData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write planet file: %w", err)
		}
	}

	return finalData, nil
}

func ParseIdentityPublic(s string) (*ZtWorldPlanetNodeIdentity, error) {
	parts := splitByColon(s)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid identity format, expected format: address:0:publicKey")
	}

	identity := &ZtWorldPlanetNodeIdentity{}

	addrBytes, err := hex.DecodeString(parts[0])
	if err != nil || len(addrBytes) != 5 {
		return nil, fmt.Errorf("invalid address (must be 10 hex chars): %w", err)
	}
	copy(identity.ZtNodeAddress[:], addrBytes)

	pubBytes, err := hex.DecodeString(parts[2])
	if err != nil || len(pubBytes) != 32 {
		return nil, fmt.Errorf("invalid public key (must be 64 hex chars): %w", err)
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

func generatePlanetID() uint64 {
	b := make([]byte, 4)
	rand.Read(b)
	return uint64(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
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
