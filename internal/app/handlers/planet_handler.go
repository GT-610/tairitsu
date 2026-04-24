/*
 * Tairitsu - A ZeroTier Network Controller Manager
 * Copyright (C) 2025 Patmeow Lab
 * SPDX-License-Identifier: GPL-3.0-only
 */

package handlers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GT-610/tairitsu/internal/mkworld"
	"github.com/gofiber/fiber/v3"
)

type GeneratePlanetRequest struct {
	IdentityPublic string   `json:"identity_public"`
	Endpoints      []string `json:"endpoints"`
	Comments       string   `json:"comments"`
	SigningKeyPath string   `json:"signing_key_path"`
}

type GeneratePlanetResponse struct {
	Message      string `json:"message"`
	PlanetData   []byte `json:"planet_data"`
	PlanetID     uint64 `json:"planet_id"`
	BirthTime    int64  `json:"birth_time"`
	DownloadName string `json:"download_name"`
}

type IdentityInfoResponse struct {
	Message        string `json:"message"`
	IdentityPublic string `json:"identity_public"`
	IdentityPath   string `json:"identity_path"`
}

type SigningKeysInfoResponse struct {
	Message         string `json:"message"`
	SigningKeyPath  string `json:"signing_key_path"`
	PreviousKeyPath string `json:"previous_key_path"`
	CurrentKeyPath  string `json:"current_key_path"`
	PreviousExists  bool   `json:"previous_exists"`
	CurrentExists   bool   `json:"current_exists"`
	Ready           bool   `json:"ready"`
}

type GenerateSigningKeysResponse struct {
	Message         string `json:"message"`
	SigningKeyPath  string `json:"signing_key_path"`
	PreviousKeyPath string `json:"previous_key_path"`
	CurrentKeyPath  string `json:"current_key_path"`
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
			Error: "identity_public is required",
		})
	}

	generatedPlanet, err := mkworld.GeneratePlanet(&mkworld.GenerateOptions{
		IdentityPublic: req.IdentityPublic,
		Endpoints:      req.Endpoints,
		Comments:       req.Comments,
		SigningKeyPath: strings.TrimSpace(req.SigningKeyPath),
	})
	if err != nil {
		switch {
		case errors.Is(err, mkworld.ErrIdentityPublicRequired),
			errors.Is(err, mkworld.ErrNoEndpoints),
			errors.Is(err, mkworld.ErrInvalidIdentity),
			errors.Is(err, mkworld.ErrInvalidEndpoint),
			errors.Is(err, mkworld.ErrDuplicateEndpoint),
			errors.Is(err, mkworld.ErrInvalidSigningKeys):
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: err.Error(),
			})
		default:
			return c.Status(500).JSON(ErrorResponse{
				Error:   "Failed to generate planet",
				Details: err.Error(),
			})
		}
	}

	return c.JSON(GeneratePlanetResponse{
		Message:      "Planet generated successfully",
		PlanetData:   generatedPlanet.PlanetData,
		PlanetID:     generatedPlanet.PlanetID,
		BirthTime:    generatedPlanet.BirthTime,
		DownloadName: "planet",
	})
}

func GetIdentityHandler(c fiber.Ctx) error {
	ztPath := c.Query("path", "/var/lib/zerotier-one")
	identityPath := filepath.Join(ztPath, "identity.public")

	identityPublic, err := os.ReadFile(identityPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":         fmt.Sprintf("identity.public not found at %s", identityPath),
				"identity_path": identityPath,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":         "Failed to read identity.public",
			"details":       err.Error(),
			"identity_path": identityPath,
		})
	}

	return c.JSON(IdentityInfoResponse{
		Message:        "Identity read successfully",
		IdentityPublic: strings.TrimSpace(string(identityPublic)),
		IdentityPath:   identityPath,
	})
}

func GetSigningKeysInfoHandler(c fiber.Ctx) error {
	ztPath := c.Query("path", "/var/lib/zerotier-one")
	prevPath := filepath.Join(ztPath, "previous.c25519")
	curPath := filepath.Join(ztPath, "current.c25519")

	_, prevErr := os.Stat(prevPath)
	_, curErr := os.Stat(curPath)
	if prevErr != nil && !os.IsNotExist(prevErr) {
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Failed to inspect previous signing key",
			Details: prevErr.Error(),
		})
	}
	if curErr != nil && !os.IsNotExist(curErr) {
		return c.Status(500).JSON(ErrorResponse{
			Error:   "Failed to inspect current signing key",
			Details: curErr.Error(),
		})
	}
	prevExists := prevErr == nil
	curExists := curErr == nil

	return c.JSON(SigningKeysInfoResponse{
		Message:         "Signing key status loaded successfully",
		SigningKeyPath:  ztPath,
		PreviousKeyPath: prevPath,
		CurrentKeyPath:  curPath,
		PreviousExists:  prevExists,
		CurrentExists:   curExists,
		Ready:           prevExists && curExists,
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

	return c.JSON(GenerateSigningKeysResponse{
		Message:         "Signing keys generated successfully",
		SigningKeyPath:  ztPath,
		PreviousKeyPath: prevPath,
		CurrentKeyPath:  curPath,
	})
}
