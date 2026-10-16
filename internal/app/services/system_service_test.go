package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock is a manually advanced time source, so cache ageing is asserted
// instead of waited for.
type fakeClock struct {
	current time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Unix(1_700_000_000, 0)}
}

func (c *fakeClock) now() time.Time {
	return c.current
}

func (c *fakeClock) advance(d time.Duration) {
	c.current = c.current.Add(d)
}

// countingCollector records how many times the service actually sampled the
// machine and can be made to fail.
type countingCollector struct {
	calls int
	err   error
}

func (c *countingCollector) collect(now func() time.Time) func() (*SystemStats, error) {
	return func() (*SystemStats, error) {
		c.calls++
		if c.err != nil {
			return nil, c.err
		}
		return &SystemStats{
			CPUUsage:    12.5,
			MemoryUsage: 40,
			Timestamp:   now().UnixMilli(),
		}, nil
	}
}

func newStatsTestService(clock *fakeClock, collector *countingCollector, expiry time.Duration) *SystemService {
	return NewSystemService(
		WithClock(clock.now),
		WithStatsCacheExpiry(expiry),
		withStatsCollector(collector.collect(clock.now)),
	)
}

func TestSystemServiceServesASecondCallFromCache(t *testing.T) {
	clock := newFakeClock()
	collector := &countingCollector{}
	service := newStatsTestService(clock, collector, time.Minute)

	first, err := service.GetSystemStats()
	require.NoError(t, err)
	require.NotNil(t, first)

	// Any movement inside the expiry window must not reach the collector.
	clock.advance(time.Minute - time.Millisecond)

	second, err := service.GetSystemStats()
	require.NoError(t, err)

	assert.Equal(t, 1, collector.calls)
	assert.Same(t, first, second, "the cached sample is handed back, not rebuilt")
	assert.Equal(t, 12.5, second.CPUUsage)
	assert.Equal(t, int64(1_700_000_000_000), second.Timestamp)
}

func TestSystemServiceResamplesOnceTheCacheExpires(t *testing.T) {
	clock := newFakeClock()
	collector := &countingCollector{}
	service := newStatsTestService(clock, collector, time.Minute)

	first, err := service.GetSystemStats()
	require.NoError(t, err)

	clock.advance(time.Minute)

	second, err := service.GetSystemStats()
	require.NoError(t, err)

	assert.Equal(t, 2, collector.calls)
	assert.NotSame(t, first, second)
	assert.Equal(t, clock.now().UnixMilli(), second.Timestamp)
}

// A failed sample must not be cached: serving a stale reading as fresh would
// report a machine state that may no longer hold, and the caller has to see the
// error on every attempt until a collection succeeds.
func TestSystemServiceDoesNotCacheAFailedSample(t *testing.T) {
	clock := newFakeClock()
	collector := &countingCollector{err: errors.New("cpu unavailable")}
	service := newStatsTestService(clock, collector, time.Minute)

	stats, err := service.GetSystemStats()
	require.Error(t, err)
	assert.Nil(t, stats)

	stats, err = service.GetSystemStats()
	require.Error(t, err)
	assert.Nil(t, stats)

	assert.Equal(t, 2, collector.calls, "a failure must not populate the cache")
}

// The service is reached from concurrent HTTP requests, so the cache has to hold
// up under the race detector.
func TestSystemServiceReadsStatsConcurrently(t *testing.T) {
	service := NewSystemService(
		WithStatsCacheExpiry(time.Minute),
		withStatsCollector(func() (*SystemStats, error) {
			return &SystemStats{Timestamp: time.Now().UnixMilli()}, nil
		}),
	)

	done := make(chan struct{})
	for range 8 {
		go func() {
			defer func() { done <- struct{}{} }()
			for range 20 {
				if _, err := service.GetSystemStats(); err != nil {
					t.Errorf("GetSystemStats returned %v", err)
					return
				}
			}
		}()
	}

	for range 8 {
		<-done
	}
}
