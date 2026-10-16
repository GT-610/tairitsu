package services

import (
	"sync"
	"time"

	"github.com/GT-610/tairitsu/internal/app/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

const defaultStatsCacheExpiry = 5 * time.Second

// SystemStats represents system resource usage statistics
type SystemStats struct {
	CPUUsage        float64 `json:"cpuUsage"`        // CPU usage percentage
	MemoryUsage     float64 `json:"memoryUsage"`     // Memory usage percentage
	Timestamp       int64   `json:"timestamp"`       // Unix timestamp in milliseconds
	OSName          string  `json:"osName"`          // Operating system name
	Platform        string  `json:"platform"`        // Platform (linux, windows, darwin)
	PlatformVersion string  `json:"platformVersion"` // Platform version
	KernelVersion   string  `json:"kernelVersion"`   // Kernel version
}

// SystemService handles system-related operations and statistics
type SystemService struct {
	statsCache  *SystemStats
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
	now         func() time.Time
	collect     func() (*SystemStats, error)
}

// SystemServiceOption customises a SystemService. The expiry, the clock and the
// collector are injectable so the cache can be exercised against a controlled
// clock and sample instead of waiting a second for a real CPU reading.
type SystemServiceOption func(*SystemService)

// WithStatsCacheExpiry replaces how long a collected sample stays fresh.
func WithStatsCacheExpiry(expiry time.Duration) SystemServiceOption {
	return func(s *SystemService) {
		s.cacheExpiry = expiry
	}
}

// WithClock replaces the time source used to stamp and age samples.
func WithClock(now func() time.Time) SystemServiceOption {
	return func(s *SystemService) {
		if now != nil {
			s.now = now
		}
	}
}

// withStatsCollector replaces the sample source. It stays unexported because
// production always reads the machine it runs on.
func withStatsCollector(collect func() (*SystemStats, error)) SystemServiceOption {
	return func(s *SystemService) {
		if collect != nil {
			s.collect = collect
		}
	}
}

// NewSystemService creates a new system service instance
func NewSystemService(options ...SystemServiceOption) *SystemService {
	service := &SystemService{
		cacheExpiry: defaultStatsCacheExpiry,
		now:         time.Now,
	}

	for _, option := range options {
		option(service)
	}

	if service.collect == nil {
		service.collect = service.collectSystemStats
	}

	return service
}

// GetSystemStats retrieves the current system statistics
// It uses a cache to avoid frequent system calls
func (s *SystemService) GetSystemStats() (*SystemStats, error) {
	if cached := s.freshCachedStats(); cached != nil {
		return cached, nil
	}

	// Cache is invalid, update it
	stats, err := s.collect()
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

// freshCachedStats returns the cached sample while it is still fresh. The
// freshness test and the read share one critical section: deciding outside the
// lock and reading inside it leaves a window in which another caller can replace
// the entry, so the caller could receive the value it just decided to discard.
func (s *SystemService) freshCachedStats() *SystemStats {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if s.statsCache == nil {
		return nil
	}

	// Check if cache has expired
	if s.now().Sub(time.UnixMilli(s.statsCache.Timestamp)) >= s.cacheExpiry {
		return nil
	}

	return s.statsCache
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

	// Get OS information
	hostInfo, err := host.Info()
	if err != nil {
		logger.Error("Failed to get OS info", zap.Error(err))
		return nil, err
	}

	return &SystemStats{
		CPUUsage:        cpuUsage,
		MemoryUsage:     memoryUsage,
		Timestamp:       s.now().UnixMilli(),
		OSName:          hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		KernelVersion:   hostInfo.KernelVersion,
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
