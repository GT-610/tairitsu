/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import "errors"

var (
	ErrInvalidIdentity        = errors.New("invalid identity format")
	ErrInvalidEndpoint        = errors.New("invalid endpoint format")
	ErrDuplicateEndpoint      = errors.New("duplicate endpoint")
	ErrDuplicateIdentity      = errors.New("duplicate root identity")
	ErrMaxEndpointsExceeded   = errors.New("endpoint count exceeds maximum")
	ErrMaxRootNodesExceeded   = errors.New("root node count exceeds maximum")
	ErrSerializedDataTooLarge = errors.New("serialized data too large")
	ErrInvalidSigningKeys     = errors.New("invalid signing keys")
	ErrNoRootNodes            = errors.New("at least one root node is required")
	ErrReservedPlanetID       = errors.New("planet id is reserved")
	ErrInvalidBirthTime       = errors.New("birth time is invalid")
	ErrUnknown                = errors.New("unknown error")
)

const (
	ZT_C25519_PUBLIC_KEY_LEN               = 64
	ZT_C25519_PRIVATE_KEY_LEN              = 64
	ZT_C25519_SIGNATURE_LEN                = 96
	ZT_WORLD_MAX_ROOTS                     = 4
	ZT_WORLD_MAX_STABLE_ENDPOINTS_PER_ROOT = 32
	ZT_WORLD_MAX_SERIALIZED_LENGTH         = ((1024 + (32 * ZT_WORLD_MAX_STABLE_ENDPOINTS_PER_ROOT)) * ZT_WORLD_MAX_ROOTS) + ZT_C25519_PUBLIC_KEY_LEN + ZT_C25519_SIGNATURE_LEN + 128
)
