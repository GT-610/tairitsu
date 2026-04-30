/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"crypto/rand"
	"crypto/sha512"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
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

func GenerateDualPair() (pub [64]byte, priv [64]byte) {
	k0pub, k0priv, _ := ed25519.GenerateKey(rand.Reader)
	var k1pub, k1priv [32]byte
	rand.Read(k1priv[:])
	curve25519.ScalarBaseMult(&k1pub, &k1priv)

	copy(pub[0:32], k1pub[:])
	copy(priv[0:32], k1priv[:])
	copy(pub[32:64], k0pub[:])
	copy(priv[32:64], k0priv[:])

	return
}
