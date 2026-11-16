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
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RootNodeConfig struct {
	IdentityPublic string   `json:"identity_public"`
	Comments       string   `json:"comments"`
	Endpoints      []string `json:"endpoints"`
}

type GenerateOptions struct {
	RootNodes       []RootNodeConfig
	SigningKeyPath  string
	PlanetID        uint64
	BirthTime       int64
	RecommendValues bool
	DownloadName    string
}

type GeneratedPlanet struct {
	PlanetID              uint64
	BirthTime             int64
	PlanetData            []byte
	DownloadName          string
	RootNodeCount         int
	EndpointCount         int
	UsedRecommendedValues bool
}

var (
	ErrIdentityPublicRequired = errors.New("identity.public is required")
	ErrNoEndpoints            = errors.New("at least one endpoint is required")
)

func GeneratePlanet(opts *GenerateOptions) (*GeneratedPlanet, error) {
	if len(opts.RootNodes) == 0 {
		return nil, ErrNoRootNodes
	}
	if len(opts.RootNodes) > ZT_WORLD_MAX_ROOTS {
		return nil, ErrMaxRootNodesExceeded
	}

	prevPub, curPub, prevPriv, curPriv, err := loadSigningKeys(opts.SigningKeyPath)
	if err != nil {
		return nil, err
	}

	planetID, birthTime, usedRecommendedValues, err := resolvePlanetMetadata(opts)
	if err != nil {
		return nil, err
	}

	nodes := make([]*ZtWorldPlanetNode, 0, len(opts.RootNodes))
	seenIdentities := make(map[string]struct{}, len(opts.RootNodes))
	totalEndpoints := 0

	for _, rootNodeConfig := range opts.RootNodes {
		if strings.TrimSpace(rootNodeConfig.IdentityPublic) == "" {
			return nil, ErrIdentityPublicRequired
		}

		identity, err := ParseIdentityPublic(rootNodeConfig.IdentityPublic)
		if err != nil {
			return nil, err
		}

		identityKey := strings.TrimSpace(rootNodeConfig.IdentityPublic)
		if _, exists := seenIdentities[identityKey]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateIdentity, identity.ZtNodeAddressString())
		}
		seenIdentities[identityKey] = struct{}{}

		endpoints, err := parseRootNodeEndpoints(rootNodeConfig.Endpoints)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, &ZtWorldPlanetNode{
			Identity:  identity,
			Endpoints: endpoints,
			Comments:  strings.TrimSpace(rootNodeConfig.Comments),
		})
		totalEndpoints += len(endpoints)

	}

	ztW := &ZtWorld{
		Type:      ZT_WORLD_TYPE_PLANET,
		ID:        ZtWorldID(planetID),
		Timestamp: uint64(birthTime),
		Nodes:     nodes,
	}

	ztW.PublicKeyMustBeSignedByNextTime = curPub

	toSignData, err := ztW.Serialize(true, [ZT_C25519_SIGNATURE_LEN]byte{})
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

	// The world just built advertises curPub as the key the next one must be signed
	// by, so the stored ring has to move on for the next publication to match.
	if strings.TrimSpace(opts.SigningKeyPath) != "" {
		prevPath, curPath := signingKeyPaths(opts.SigningKeyPath)
		if err := advanceSigningKeys(prevPath, curPath, curPub, curPriv); err != nil {
			return nil, err
		}
	}

	return &GeneratedPlanet{
		PlanetID:              planetID,
		BirthTime:             birthTime,
		PlanetData:            finalData,
		DownloadName:          normalizeDownloadName(opts.DownloadName),
		RootNodeCount:         len(nodes),
		EndpointCount:         totalEndpoints,
		UsedRecommendedValues: usedRecommendedValues,
	}, nil
}

func parseRootNodeEndpoints(values []string) ([]*ZtNodeInetAddr, error) {
	endpointValues := normalizeEndpoints(values)
	if len(endpointValues) == 0 {
		return nil, ErrNoEndpoints
	}

	endpoints := make([]*ZtNodeInetAddr, 0, len(endpointValues))
	seenEndpoints := make(map[string]struct{}, len(endpointValues))
	for _, epStr := range endpointValues {
		ep := &ZtNodeInetAddr{}
		if err := ep.FromString(epStr); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEndpoint, epStr)
		}
		endpointKey := ep.IP.String() + "/" + fmt.Sprintf("%d", ep.Port)
		if _, exists := seenEndpoints[endpointKey]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateEndpoint, epStr)
		}
		seenEndpoints[endpointKey] = struct{}{}
		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

func resolvePlanetMetadata(opts *GenerateOptions) (planetID uint64, birthTime int64, usedRecommendedValues bool, err error) {
	if opts.RecommendValues {
		planetID, err = generatePlanetID()
		if err != nil {
			err = fmt.Errorf("failed to generate planet id: %w", err)
			return
		}
		birthTime = time.Now().UnixMilli()
		usedRecommendedValues = true
		return
	}

	if opts.PlanetID == 0 {
		err = ErrReservedPlanetID
		return
	}
	if opts.PlanetID == uint64(ZT_WORLD_ID_EARTH) || opts.PlanetID == uint64(ZT_WORLD_ID_MARS) {
		err = ErrReservedPlanetID
		return
	}
	if opts.BirthTime <= 1567191349589 {
		err = ErrInvalidBirthTime
		return
	}

	planetID = opts.PlanetID
	birthTime = opts.BirthTime
	return
}

func normalizeDownloadName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "planet"
	}
	return trimmed
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
	if err != nil || len(pubBytes) != ZT_C25519_PUBLIC_KEY_LEN {
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

func loadSigningKeys(signingKeyPath string) (prevPub, curPub [ZT_C25519_PUBLIC_KEY_LEN]byte, prevPriv, curPriv [ZT_C25519_PRIVATE_KEY_LEN]byte, err error) {
	if strings.TrimSpace(signingKeyPath) == "" {
		// No stored keys, so this is a throwaway pair. The two halves are still
		// generated separately: previous signs the world being published and
		// current is the successor it names.
		if prevPub, prevPriv, err = GenerateDualPair(); err != nil {
			return
		}
		curPub, curPriv, err = GenerateDualPair()
		return
	}

	prevPath, curPath := signingKeyPaths(signingKeyPath)
	prevPub, curPub, prevPriv, curPriv, err = ReadSigningKeys(prevPath, curPath)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrInvalidSigningKeys, err)
	}
	return
}

// signingKeyPaths returns the two files a signing key directory holds.
func signingKeyPaths(dir string) (prevPath, curPath string) {
	return filepath.Join(dir, "previous.c25519"), filepath.Join(dir, "current.c25519")
}

// advanceSigningKeys moves the ring on so the next publication is signed by the key
// this one advertised.
//
// A planet names current.c25519 as the key the next update must be signed by, so a
// second publication signed by previous.c25519 again would be rejected by every node
// that accepted the first. Promoting the advertised key to signer, and generating a
// fresh successor, is what makes consecutive publications a chain.
//
// The signer is written first. If the second write then fails, both files hold the
// promoted key: the next publication signs with it and advertises it, which is
// self-consistent and merely postpones the rotation, rather than naming a successor
// the next publication cannot sign with.
func advanceSigningKeys(prevPath, curPath string, nextPrevPub [ZT_C25519_PUBLIC_KEY_LEN]byte, nextPrevPriv [ZT_C25519_PRIVATE_KEY_LEN]byte) error {
	nextCurPub, nextCurPriv, err := GenerateDualPair()
	if err != nil {
		return err
	}

	if err := writeKeyFile(prevPath, nextPrevPub, nextPrevPriv); err != nil {
		return fmt.Errorf("failed to promote the published signing key: %w", err)
	}
	if err := writeKeyFile(curPath, nextCurPub, nextCurPriv); err != nil {
		return fmt.Errorf("failed to write the next signing key: %w", err)
	}

	return nil
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

// CreateSigningKeys writes the rotation pair ZeroTier expects.
//
// previous.c25519 signs the world being published; current.c25519 is the
// successor whose signature the next world must carry. They are two separate
// pairs on purpose — writing one pair to both paths, as this used to, produced a
// rotation in which the successor was the same key, so the rollover the model
// exists for could never happen.
//
// An existing pair is refused rather than replaced. A planet that is already
// deployed names the stored current key as the one its next update must be signed
// by, so overwriting it silently breaks that update chain; starting a new identity
// has to be a deliberate act, which removing the files is.
func CreateSigningKeys(prevPath, curPath string) error {
	for _, path := range []string{prevPath, curPath} {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s", ErrSigningKeysExist, path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to inspect %s: %w", path, err)
		}
	}

	prevPub, prevPriv, err := GenerateDualPair()
	if err != nil {
		return err
	}

	curPub, curPriv, err := GenerateDualPair()
	if err != nil {
		return err
	}

	if err := EnsureDirectory(prevPath); err != nil {
		return fmt.Errorf("failed to prepare previous key directory: %w", err)
	}
	if err := EnsureDirectory(curPath); err != nil {
		return fmt.Errorf("failed to prepare current key directory: %w", err)
	}

	if err := writeKeyFile(prevPath, prevPub, prevPriv); err != nil {
		return fmt.Errorf("failed to write previous key: %w", err)
	}
	if err := writeKeyFile(curPath, curPub, curPriv); err != nil {
		return fmt.Errorf("failed to write current key: %w", err)
	}

	return nil
}

func writeKeyFile(path string, pub [ZT_C25519_PUBLIC_KEY_LEN]byte, priv [ZT_C25519_PRIVATE_KEY_LEN]byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ensure permissions are 0600 even for pre-existing files (O_CREATE|O_TRUNC
	// does not change the mode of an already-existing file on Linux).
	if err := file.Chmod(0600); err != nil {
		return err
	}

	if _, err := file.Write(pub[:]); err != nil {
		return err
	}
	if _, err := file.Write(priv[:]); err != nil {
		return err
	}

	return nil
}

func ReadSigningKeys(prevPath, curPath string) (prevPub, curPub [ZT_C25519_PUBLIC_KEY_LEN]byte, prevPriv, curPriv [ZT_C25519_PRIVATE_KEY_LEN]byte, err error) {
	prevData, err := os.ReadFile(prevPath)
	if err != nil {
		return
	}
	curData, err := os.ReadFile(curPath)
	if err != nil {
		return
	}

	if len(prevData) != ZT_C25519_PUBLIC_KEY_LEN+ZT_C25519_PRIVATE_KEY_LEN || len(curData) != ZT_C25519_PUBLIC_KEY_LEN+ZT_C25519_PRIVATE_KEY_LEN {
		err = fmt.Errorf("invalid key file length")
		return
	}

	copy(prevPub[:], prevData[:ZT_C25519_PUBLIC_KEY_LEN])
	copy(prevPriv[:], prevData[ZT_C25519_PUBLIC_KEY_LEN:ZT_C25519_PUBLIC_KEY_LEN+ZT_C25519_PRIVATE_KEY_LEN])
	copy(curPub[:], curData[:ZT_C25519_PUBLIC_KEY_LEN])
	copy(curPriv[:], curData[ZT_C25519_PUBLIC_KEY_LEN:ZT_C25519_PUBLIC_KEY_LEN+ZT_C25519_PRIVATE_KEY_LEN])

	return
}

func EnsureDirectory(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
