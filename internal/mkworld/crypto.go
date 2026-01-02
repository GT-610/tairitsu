/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/salsa20"
)

var ErrSigningFailed = errors.New("failed to sign message")

func GenerateSigningKeyPair() (pub [32]byte, priv [32]byte) {
	rand.Read(priv[:])
	curve25519.ScalarBaseMult(&pub, &priv)
	return
}

func SignMessage(pub [32]byte, priv [32]byte, msg []byte) ([64]byte, error) {
	var sigBuf [96]byte

	h := sha512.New()
	h.Write(msg)
	s512 := h.Sum(nil)
	copy(sigBuf[64:], s512[:32])

	goPrivK := make([]byte, 64)
	copy(goPrivK[:32], priv[:])
	copy(goPrivK[32:], pub[:])
	sigData := ed25519.Sign(goPrivK, s512[:32])
	copy(sigBuf[:64], sigData)

	var finalSig [64]byte
	copy(finalSig[:], sigBuf[0:64])
	copy(finalSig[64:], sigBuf[64:96])

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

func ComputeZeroTierIdentityMemoryHardHash(publicKey []byte) []byte {
	s512 := sha512.Sum512(publicKey)

	const ztIdentityGenMemory = 2097152
	var genmem [ztIdentityGenMemory]byte
	var s20key [32]byte
	var s20ctr [16]byte
	var s20ctri uint64
	copy(s20key[:], s512[0:32])
	copy(s20ctr[0:8], s512[32:40])

	salsa20.XORKeyStream(genmem[0:64], genmem[0:64], s20ctr[:], &s20key)
	s20ctri++
	for i := 64; i < ztIdentityGenMemory; i += 64 {
		binary.LittleEndian.PutUint64(s20ctr[8:16], s20ctri)
		salsa20.XORKeyStream(genmem[i:i+64], genmem[i-64:i], s20ctr[:], &s20key)
		s20ctri++
	}

	var tmp [8]byte
	for i := 0; i < ztIdentityGenMemory; {
		idx1 := uint(binary.BigEndian.Uint64(genmem[i:])&7) * 8
		i += 8
		idx2 := (uint(binary.BigEndian.Uint64(genmem[i:])) % uint(ztIdentityGenMemory/8)) * 8
		i += 8
		gm := genmem[idx2 : idx2+8]
		d := s512[idx1 : idx1+8]
		copy(tmp[:], gm)
		copy(gm, d)
		copy(d, tmp[:])
		binary.LittleEndian.PutUint64(s20ctr[8:16], s20ctri)
		salsa20.XORKeyStream(s512[:], s512[:], s20ctr[:], &s20key)
		s20ctri++
	}

	return s512[:]
}
