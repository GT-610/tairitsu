/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package mkworld

import (
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

type ZtWorldType uint8

const (
	ZT_WORLD_TYPE_NULL ZtWorldType = iota
	ZT_WORLD_TYPE_PLANET
	ZT_WORLD_TYPE_MOON = 127
)

type ZtWorldID uint64

const (
	ZT_WORLD_ID_EARTH ZtWorldID = 149604618
	ZT_WORLD_ID_MARS  ZtWorldID = 227883110
)

type ZtNodeInetAddr struct {
	IP   net.IP
	Port uint16
}

func (a *ZtNodeInetAddr) FromString(ipport string) error {
	parts := strings.Split(ipport, "/")
	if len(parts) != 2 {
		return ErrInvalidEndpoint
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return ErrInvalidEndpoint
	}

	port, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return ErrInvalidEndpoint
	}

	a.IP = ip
	a.Port = uint16(port)
	return nil
}

func (a *ZtNodeInetAddr) Serialize() ([]byte, error) {
	var buf []byte

	if a.IP.To4() != nil {
		buf = append(buf, 4)
		buf = append(buf, a.IP.To4()...)
	} else {
		buf = append(buf, 6)
		buf = append(buf, a.IP.To16()...)
	}

	buf = binary.BigEndian.AppendUint16(buf, a.Port)
	return buf, nil
}

type ZtWorldPlanetNodeIdentity struct {
	ZtNodeAddress [5]byte
	PublicKey     [32]byte
}

func (id *ZtWorldPlanetNodeIdentity) FromString(data string) error {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return ErrInvalidIdentity
	}

	addrBytes, err := hex.DecodeString(parts[0])
	if err != nil || len(addrBytes) != 5 {
		return ErrInvalidIdentity
	}
	copy(id.ZtNodeAddress[:], addrBytes)

	pubBytes, err := hex.DecodeString(parts[2])
	if err != nil || len(pubBytes) != 32 {
		return ErrInvalidIdentity
	}
	copy(id.PublicKey[:], pubBytes)

	return nil
}

func (id *ZtWorldPlanetNodeIdentity) Serialize() ([]byte, error) {
	var buf []byte
	buf = append(buf, id.ZtNodeAddress[:]...)
	buf = append(buf, 0)
	buf = append(buf, id.PublicKey[:]...)
	buf = append(buf, 0)
	return buf, nil
}

type ZtWorldPlanetNode struct {
	Identity  *ZtWorldPlanetNodeIdentity
	Endpoints []*ZtNodeInetAddr
}

func (n *ZtWorldPlanetNode) Serialize() ([]byte, error) {
	var buf []byte

	idData, err := n.Identity.Serialize()
	if err != nil {
		return nil, err
	}
	buf = append(buf, idData...)

	buf = append(buf, uint8(len(n.Endpoints)))

	for _, ep := range n.Endpoints {
		epData, err := ep.Serialize()
		if err != nil {
			return nil, err
		}
		buf = append(buf, epData...)
	}

	if len(n.Endpoints) > ZT_WORLD_MAX_STABLE_ENDPOINTS_PER_ROOT {
		return nil, ErrMaxEndpointsExceeded
	}

	return buf, nil
}

type ZtWorld struct {
	Type                            ZtWorldType
	ID                              ZtWorldID
	Timestamp                       uint64
	PublicKeyMustBeSignedByNextTime [32]byte
	Nodes                           []*ZtWorldPlanetNode
}

func (w *ZtWorld) Serialize(forSign bool, c25519sig [64]byte) ([]byte, error) {
	var buf []byte

	if forSign {
		buf = binary.BigEndian.AppendUint64(buf, 0x7f7f7f7f7f7f7f7f)
	}

	buf = append(buf, byte(w.Type))
	buf = binary.BigEndian.AppendUint64(buf, uint64(w.ID))
	buf = binary.BigEndian.AppendUint64(buf, w.Timestamp)
	buf = append(buf, w.PublicKeyMustBeSignedByNextTime[:]...)

	if !forSign && c25519sig != [64]byte{} {
		buf = append(buf, c25519sig[:]...)
	}

	buf = append(buf, uint8(len(w.Nodes)))

	if len(w.Nodes) > ZT_WORLD_MAX_ROOTS {
		return nil, ErrSerializedDataTooLarge
	}

	for _, n := range w.Nodes {
		nBytes, err := n.Serialize()
		if err != nil {
			return nil, err
		}
		buf = append(buf, nBytes...)
	}

	if w.Type == ZT_WORLD_TYPE_MOON {
		buf = binary.BigEndian.AppendUint16(buf, 0)
	}

	if forSign {
		buf = binary.BigEndian.AppendUint64(buf, 0xf7f7f7f7f7f7f7f7)
	}

	if len(buf) > ZT_WORLD_MAX_SERIALIZED_LENGTH {
		return nil, ErrSerializedDataTooLarge
	}

	return buf, nil
}
