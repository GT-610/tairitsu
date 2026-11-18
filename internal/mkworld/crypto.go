/*
 * Tempest-ZT - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)

func SignMessage(pub [ZT_C25519_PUBLIC_KEY_LEN]byte, priv [ZT_C25519_PRIVATE_KEY_LEN]byte, msg []byte) ([ZT_C25519_SIGNATURE_LEN]byte, error) {
	var sigBuf [ZT_C25519_SIGNATURE_LEN]byte

	h := sha512.New()
	h.Write(msg)
	s512 := h.Sum(nil)
	copy(sigBuf[64:], s512[:32])

	goPrivK := make([]byte, 64)
	copy(goPrivK[:32], priv[32:64])
	copy(goPrivK[32:], pub[32:64])
	sigData := ed25519.Sign(goPrivK, s512[:32])
	copy(sigBuf[:64], sigData)

	var finalSig [ZT_C25519_SIGNATURE_LEN]byte
	copy(finalSig[:], sigBuf[:])

	return finalSig, nil
}

// GenerateDualPair produces one ZeroTier identity pair: a Curve25519 half and an
// Ed25519 half.
//
// A failure from the random source is returned rather than ignored. Ignoring it
// would leave the Curve25519 private half at its zero value, which is a key
// anyone can reproduce.
func GenerateDualPair() (pub [64]byte, priv [64]byte, err error) {
	k0pub, k0priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return pub, priv, fmt.Errorf("failed to generate the ed25519 key: %w", err)
	}

	var k1pub, k1priv [32]byte
	if _, err := io.ReadFull(rand.Reader, k1priv[:]); err != nil {
		return pub, priv, fmt.Errorf("failed to generate the curve25519 key: %w", err)
	}
	curve25519.ScalarBaseMult(&k1pub, &k1priv)

	copy(pub[0:32], k1pub[:])
	copy(priv[0:32], k1priv[:])
	copy(pub[32:64], k0pub[:])
	copy(priv[32:64], k0priv[:])

	return pub, priv, nil
}
