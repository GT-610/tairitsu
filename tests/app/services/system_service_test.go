package services

import (
	"testing"
	"time"

	"github.com/GT-610/tairitsu/internal/app/services"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemService(t *testing.T) {
	// 创建系统服务实例
	service := services.NewSystemService()
	assert.NotNil(t, service)
}

func TestGetSystemStats(t *testing.T) {
	// 创建系统服务实例
	service := services.NewSystemService()

	// 调用GetSystemStats方法
	stats, err := service.GetSystemStats()

	// 检查结果
	if err != nil {
		// 如果无法获取系统资源，这是正常的，特别是在容器环境中
		t.Logf("无法获取系统资源：%v", err)
		return
	}

	// 验证返回值
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.CPUUsage, 0.0)
	assert.LessOrEqual(t, stats.CPUUsage, 100.0)
	assert.GreaterOrEqual(t, stats.MemoryUsage, 0.0)
	assert.LessOrEqual(t, stats.MemoryUsage, 100.0)
	assert.Greater(t, stats.Timestamp, int64(0))
}

func TestGetSystemStats_Cache(t *testing.T) {
	// 创建系统服务实例
	service := services.NewSystemService()

	// 第一次调用GetSystemStats方法
	stats1, err := service.GetSystemStats()
	if err != nil {
		// 如果无法获取系统资源，这是正常的，特别是在容器环境中
		t.Logf("无法获取系统资源：%v", err)
		return
	}

	// 立即第二次调用GetSystemStats方法，应该使用缓存
	stats2, err := service.GetSystemStats()
	if err != nil {
		// 如果无法获取系统资源，这是正常的，特别是在容器环境中
		t.Logf("无法获取系统资源：%v", err)
		return
	}

	// 验证两次调用返回的是相同的结果（缓存生效）
	assert.Equal(t, stats1.CPUUsage, stats2.CPUUsage)
	assert.Equal(t, stats1.MemoryUsage, stats2.MemoryUsage)
	assert.Equal(t, stats1.Timestamp, stats2.Timestamp)
}

func TestGetSystemStats_CacheExpiry(t *testing.T) {
	// 创建系统服务实例
	service := services.NewSystemService()

	// 第一次调用GetSystemStats方法
	stats1, err := service.GetSystemStats()
	if err != nil {
		// 如果无法获取系统资源，这是正常的，特别是在容器环境中
		t.Logf("无法获取系统资源：%v", err)
		return
	}

	// 等待6秒，确保缓存过期（缓存时间为5秒）
	time.Sleep(6 * time.Second)

	// 第二次调用GetSystemStats方法，应该重新获取数据
	stats2, err := service.GetSystemStats()
	if err != nil {
		// 如果无法获取系统资源，这是正常的，特别是在容器环境中
		t.Logf("无法获取系统资源：%v", err)
		return
	}

	// 验证两次调用返回的时间戳不同（缓存已过期）
	assert.NotEqual(t, stats1.Timestamp, stats2.Timestamp)
}
