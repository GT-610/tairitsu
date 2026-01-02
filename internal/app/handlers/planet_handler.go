/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GT-610/tairitsu/internal/mkworld"
	"github.com/gofiber/fiber/v3"
)

type GeneratePlanetRequest struct {
	IdentityPublic string   `json:"identityPublic"`
	Endpoints      []string `json:"endpoints"`
	Comments       string   `json:"comments"`
	OutputPath     string   `json:"outputPath,omitempty"`
}

type GeneratePlanetResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	PlanetData []byte `json:"planetData,omitempty"`
	PlanetID   uint64 `json:"planetId"`
	BirthTime  int64  `json:"birthTime"`
	CHeader    string `json:"cHeader,omitempty"`
}

type IdentityInfoResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	IdentityPublic string `json:"identityPublic,omitempty"`
	IdentityPath   string `json:"identityPath,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func GeneratePlanetHandler(c fiber.Ctx) error {
	var req GeneratePlanetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
	}

	if req.IdentityPublic == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "identityPublic is required",
		})
	}

	if len(req.Endpoints) == 0 {
		return c.Status(400).JSON(ErrorResponse{
			Error: "at least one endpoint is required",
		})
	}

	planetData, err := mkworld.GeneratePlanet(&mkworld.GenerateOptions{
		IdentityPublic: req.IdentityPublic,
		Endpoints:      req.Endpoints,
		Comments:       req.Comments,
		OutputPath:     req.OutputPath,
	})
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Failed to generate planet",
			Details: err.Error(),
		})
	}

	cHeader := generateCHeader(planetData)

	return c.JSON(GeneratePlanetResponse{
		Success:    true,
		Message:    "Planet generated successfully",
		PlanetData: planetData,
		PlanetID:   0,
		BirthTime:  0,
		CHeader:    cHeader,
	})
}

func GetIdentityHandler(c fiber.Ctx) error {
	ztPath := c.Query("path", "/var/lib/zerotier-one")
	identityPath := filepath.Join(ztPath, "identity.public")

	identityPublic, err := os.ReadFile(identityPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(IdentityInfoResponse{
				Success:      false,
				Message:      fmt.Sprintf("identity.public not found at %s", identityPath),
				IdentityPath: identityPath,
			})
		}
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Failed to read identity.public",
			Details: err.Error(),
		})
	}

	return c.JSON(IdentityInfoResponse{
		Success:        true,
		Message:        "Identity read successfully",
		IdentityPublic: strings.TrimSpace(string(identityPublic)),
		IdentityPath:   identityPath,
	})
}

func GenerateSigningKeysHandler(c fiber.Ctx) error {
	ztPath := c.Query("path", "/var/lib/zerotier-one")
	prevPath := filepath.Join(ztPath, "previous.c25519")
	curPath := filepath.Join(ztPath, "current.c25519")

	err := mkworld.CreateSigningKeys(prevPath, curPath)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Failed to generate signing keys",
			Details: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Signing keys generated successfully",
		"previousKey": prevPath,
		"currentKey":  curPath,
	})
}

func generateCHeader(data []byte) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("#define ZT_DEFAULT_WORLD_LENGTH %d\n", len(data)))
	sb.WriteString("static const unsigned char ZT_DEFAULT_WORLD[ZT_DEFAULT_WORLD_LENGTH] = {")
	for i, v := range data {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("0x%02x", v))
	}
	sb.WriteString("};\n")
	return sb.String()
}
