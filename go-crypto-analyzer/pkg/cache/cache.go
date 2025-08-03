package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// OHLCVCache 缓存管理器
type OHLCVCache struct {
	mu        sync.RWMutex
	memory    map[string]*CachedData
	cacheDir  string
	ttl       time.Duration
}

// CachedData 缓存数据结构
type CachedData struct {
	Symbol    string         `json:"symbol"`
	Interval  string         `json:"interval"`
	Data      []types.OHLCV  `json:"data"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// NewOHLCVCache 创建新的缓存管理器
func NewOHLCVCache(cacheDir string, ttl time.Duration) *OHLCVCache {
	if cacheDir == "" {
		cacheDir = ".cache"
	}
	
	// 创建缓存目录
	os.MkdirAll(cacheDir, 0755)
	
	return &OHLCVCache{
		memory:   make(map[string]*CachedData),
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

// generateKey 生成缓存键
func (c *OHLCVCache) generateKey(symbol, interval string) string {
	return fmt.Sprintf("%s_%s", symbol, interval)
}

// Get 获取缓存数据
func (c *OHLCVCache) Get(symbol, interval string) ([]types.OHLCV, bool) {
	key := c.generateKey(symbol, interval)
	
	// 先从内存缓存查找
	c.mu.RLock()
	cached, exists := c.memory[key]
	c.mu.RUnlock()
	
	if exists && time.Since(cached.UpdatedAt) < c.ttl {
		return cached.Data, true
	}
	
	// 如果内存中没有，尝试从文件加载
	cached, err := c.loadFromFile(key)
	if err == nil && time.Since(cached.UpdatedAt) < c.ttl {
		// 加载到内存
		c.mu.Lock()
		c.memory[key] = cached
		c.mu.Unlock()
		return cached.Data, true
	}
	
	return nil, false
}

// Set 设置缓存数据
func (c *OHLCVCache) Set(symbol, interval string, data []types.OHLCV) error {
	key := c.generateKey(symbol, interval)
	
	cached := &CachedData{
		Symbol:    symbol,
		Interval:  interval,
		Data:      data,
		UpdatedAt: time.Now(),
	}
	
	// 保存到内存
	c.mu.Lock()
	c.memory[key] = cached
	c.mu.Unlock()
	
	// 异步保存到文件
	go c.saveToFile(key, cached)
	
	return nil
}

// Update 更新缓存（只获取新数据）
func (c *OHLCVCache) Update(symbol, interval string, newData []types.OHLCV) error {
	key := c.generateKey(symbol, interval)
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	existing, exists := c.memory[key]
	if !exists {
		// 如果没有缓存，直接设置
		c.memory[key] = &CachedData{
			Symbol:    symbol,
			Interval:  interval,
			Data:      newData,
			UpdatedAt: time.Now(),
		}
		return nil
	}
	
	// 合并数据，去重
	merged := c.mergeData(existing.Data, newData)
	
	existing.Data = merged
	existing.UpdatedAt = time.Now()
	
	// 异步保存
	go c.saveToFile(key, existing)
	
	return nil
}

// mergeData 合并数据，去重并排序
func (c *OHLCVCache) mergeData(existing, newData []types.OHLCV) []types.OHLCV {
	// 使用map去重
	dataMap := make(map[int64]types.OHLCV)
	
	// 先添加现有数据
	for _, d := range existing {
		dataMap[d.Time.Unix()] = d
	}
	
	// 添加新数据（会覆盖相同时间的旧数据）
	for _, d := range newData {
		dataMap[d.Time.Unix()] = d
	}
	
	// 转换回切片并排序
	result := make([]types.OHLCV, 0, len(dataMap))
	for _, d := range dataMap {
		result = append(result, d)
	}
	
	// 按时间排序（使用标准库的快速排序）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.Before(result[j].Time)
	})
	
	// 只保留最新的数据（例如最多1000条）
	if len(result) > 1000 {
		result = result[len(result)-1000:]
	}
	
	return result
}

// GetLatestTime 获取缓存中最新数据的时间
func (c *OHLCVCache) GetLatestTime(symbol, interval string) (time.Time, bool) {
	data, exists := c.Get(symbol, interval)
	if !exists || len(data) == 0 {
		return time.Time{}, false
	}
	
	return data[len(data)-1].Time, true
}

// loadFromFile 从文件加载缓存
func (c *OHLCVCache) loadFromFile(key string) (*CachedData, error) {
	filename := filepath.Join(c.cacheDir, key+".json")
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var cached CachedData
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	
	return &cached, nil
}

// saveToFile 保存缓存到文件
func (c *OHLCVCache) saveToFile(key string, data *CachedData) error {
	filename := filepath.Join(c.cacheDir, key+".json")
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, jsonData, 0644)
}

// Clear 清除指定缓存
func (c *OHLCVCache) Clear(symbol, interval string) {
	key := c.generateKey(symbol, interval)
	
	c.mu.Lock()
	delete(c.memory, key)
	c.mu.Unlock()
	
	// 删除文件
	filename := filepath.Join(c.cacheDir, key+".json")
	os.Remove(filename)
}

// ClearAll 清除所有缓存
func (c *OHLCVCache) ClearAll() error {
	c.mu.Lock()
	c.memory = make(map[string]*CachedData)
	c.mu.Unlock()
	
	// 删除所有缓存文件
	return os.RemoveAll(c.cacheDir)
}

// Stats 获取缓存统计信息
func (c *OHLCVCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	stats := map[string]interface{}{
		"memory_items": len(c.memory),
		"cache_dir":    c.cacheDir,
		"ttl":          c.ttl.String(),
	}
	
	// 计算总数据点数
	totalPoints := 0
	for _, cached := range c.memory {
		totalPoints += len(cached.Data)
	}
	stats["total_data_points"] = totalPoints
	
	return stats
}