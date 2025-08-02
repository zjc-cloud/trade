package utils

import (
	"fmt"
	"time"
)

// Timer 性能计时器
type Timer struct {
	name  string
	start time.Time
}

// NewTimer 创建新的计时器
func NewTimer(name string) *Timer {
	return &Timer{
		name:  name,
		start: time.Now(),
	}
}

// Stop 停止计时并打印耗时
func (t *Timer) Stop() {
	duration := time.Since(t.start)
	if duration > time.Second {
		fmt.Printf("⏱️  %s 耗时: %.2f秒\n", t.name, duration.Seconds())
	} else if duration > time.Millisecond {
		fmt.Printf("⏱️  %s 耗时: %dms\n", t.name, duration.Milliseconds())
	}
}

// MemoryStats 内存统计
type MemoryStats struct {
	dataPoints int
	symbols    int
}

// EstimateMemoryUsage 估算内存使用
func EstimateMemoryUsage(dataPoints, symbols int) string {
	// 每个数据点大约需要：
	// OHLCV: 5 * 8 bytes = 40 bytes
	// 指标计算中间结果: ~200 bytes
	bytesPerPoint := 240
	totalBytes := dataPoints * symbols * bytesPerPoint
	
	if totalBytes > 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(totalBytes)/(1024*1024))
	} else if totalBytes > 1024 {
		return fmt.Sprintf("%.2f KB", float64(totalBytes)/1024)
	}
	return fmt.Sprintf("%d bytes", totalBytes)
}