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
	ErrMaxEndpointsExceeded   = errors.New("endpoint count exceeds maximum")
	ErrSerializedDataTooLarge = errors.New("serialized data too large")
	ErrUnknown                = errors.New("unknown error")
)

const (
	ZT_C25519_PUBLIC_KEY_LEN            = 32
	ZT_C25519_PRIVATE_KEY_LEN           = 32
	ZT_C25519_SIGNATURE_LEN             = 64
	ZT_WORLD_MAX_ROOTS                  = 4
	ZT_WORLD_MAX_STABLE_ENDPOINTS_PER_ROOT = 32
	ZT_WORLD_MAX_SERIALIZED_LENGTH      = ((1024 + (32 * ZT_WORLD_MAX_STABLE_ENDPOINTS_PER_ROOT)) * ZT_WORLD_MAX_ROOTS) + ZT_C25519_PUBLIC_KEY_LEN + ZT_C25519_SIGNATURE_LEN + 128
)
