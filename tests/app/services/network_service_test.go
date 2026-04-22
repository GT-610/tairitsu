package services

import (
	"testing"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewNetworkService(t *testing.T) {
	// Act
	networkService := services.NewNetworkService(nil, nil)

	// Assert
	assert.NotNil(t, networkService)
}

func TestImportableNetworkSummariesStartAsEmptySlice(t *testing.T) {
	summaries := make([]services.ImportableNetworkSummary, 0)

	assert.NotNil(t, summaries)
	assert.Len(t, summaries, 0)
}
