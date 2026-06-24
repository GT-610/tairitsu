package handlers

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxNetworkNameLen        = 128
	maxNetworkDescriptionLen = 1024
	maxMemberNameLen         = 128
)

func validateNetworkID(id string) error {
	id = strings.TrimSpace(id)
	if len(id) != 16 {
		return fmt.Errorf("network ID must be 16 hexadecimal characters")
	}
	if _, err := hex.DecodeString(id); err != nil {
		return fmt.Errorf("network ID must be a valid hexadecimal string")
	}
	return nil
}

func validateMemberID(id string) error {
	id = strings.TrimSpace(id)
	if len(id) != 10 {
		return fmt.Errorf("member ID must be 10 hexadecimal characters")
	}
	if _, err := hex.DecodeString(id); err != nil {
		return fmt.Errorf("member ID must be a valid hexadecimal string")
	}
	return nil
}

func validateNetworkName(name string) error {
	if utf8.RuneCountInString(name) > maxNetworkNameLen {
		return fmt.Errorf("network name must be %d characters or fewer", maxNetworkNameLen)
	}
	return nil
}

func validateNetworkDescription(desc string) error {
	if utf8.RuneCountInString(desc) > maxNetworkDescriptionLen {
		return fmt.Errorf("network description must be %d characters or fewer", maxNetworkDescriptionLen)
	}
	return nil
}

func validateMemberName(name string) error {
	if utf8.RuneCountInString(name) > maxMemberNameLen {
		return fmt.Errorf("member name must be %d characters or fewer", maxMemberNameLen)
	}
	return nil
}
