package services

import (
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

// SystemStats represents system resource usage statistics
type SystemStats struct {
	CPUUsage    float64 `json:"cpuUsage"`    // CPU usage percentage
	MemoryUsage float64 `json:"memoryUsage"` // Memory usage percentage
	Timestamp   int64   `json:"timestamp"`   // Unix timestamp in milliseconds
}

// SystemService handles system-related operations and statistics
type SystemService struct {
	statsCache  *SystemStats
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
}

// NewSystemService creates a new system service instance
func NewSystemService() *SystemService {
	return &SystemService{
		cacheExpiry: 5 * time.Second, // Cache expiry time: 5 seconds
	}
}

// GetSystemStats retrieves the current system statistics
// It uses a cache to avoid frequent system calls
func (s *SystemService) GetSystemStats() (*SystemStats, error) {
	// Check if cache is valid
	if s.isCacheValid() {
		s.cacheMutex.RLock()
		defer s.cacheMutex.RUnlock()
		return s.statsCache, nil
	}

	// Cache is invalid, update it
	stats, err := s.collectSystemStats()
	if err != nil {
		logger.Error("Failed to collect system stats", zap.Error(err))
		return nil, err
	}

	// Update cache
	s.cacheMutex.Lock()
	s.statsCache = stats
	s.cacheMutex.Unlock()

	return stats, nil
}

// isCacheValid checks if the cache is still valid
func (s *SystemService) isCacheValid() bool {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if s.statsCache == nil {
		return false
	}

	// Check if cache has expired
	cacheTime := time.UnixMilli(s.statsCache.Timestamp)
	return time.Since(cacheTime) < s.cacheExpiry
}

// collectSystemStats collects system statistics from the OS
func (s *SystemService) collectSystemStats() (*SystemStats, error) {
	// Get CPU usage
	cpuUsage, err := s.getCPUUsage()
	if err != nil {
		logger.Error("Failed to get CPU usage", zap.Error(err))
		return nil, err
	}

	// Get memory usage
	memoryUsage, err := s.getMemoryUsage()
	if err != nil {
		logger.Error("Failed to get memory usage", zap.Error(err))
		return nil, err
	}

	return &SystemStats{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		Timestamp:   time.Now().UnixMilli(),
	}, nil
}

// getCPUUsage retrieves the current CPU usage percentage
func (s *SystemService) getCPUUsage() (float64, error) {
	// Get CPU usage over 1 second interval
	// Using 1-second interval for more accurate reading
	percentages, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return 0, err
	}

	if len(percentages) == 0 {
		return 0, nil
	}

	return percentages[0], nil
}

// getMemoryUsage retrieves the current memory usage percentage
func (s *SystemService) getMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return vmStat.UsedPercent, nil
}
