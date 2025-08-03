package data

import (
	"fmt"
	"time"

	"github.com/zjc/go-crypto-analyzer/pkg/cache"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// CachedFetcher 带缓存的数据获取器
type CachedFetcher struct {
	fetcher Fetcher
	cache   *cache.OHLCVCache
}

// NewCachedFetcher 创建带缓存的数据获取器
func NewCachedFetcher(fetcher Fetcher, cacheDir string, ttl time.Duration) *CachedFetcher {
	return &CachedFetcher{
		fetcher: fetcher,
		cache:   cache.NewOHLCVCache(cacheDir, ttl),
	}
}

// FetchOHLCV 获取K线数据（优先使用缓存）
func (cf *CachedFetcher) FetchOHLCV(symbol string, interval string, limit int) ([]types.OHLCV, error) {
	// 检查缓存
	cachedData, exists := cf.cache.Get(symbol, interval)
	
	if exists && len(cachedData) >= limit {
		// 缓存数据足够，直接返回最新的数据
		start := len(cachedData) - limit
		if start < 0 {
			start = 0
		}
		return cachedData[start:], nil
	}
	
	// 缓存不存在或数据不够，需要获取新数据
	if !exists || len(cachedData) == 0 {
		// 完全没有缓存，获取全部数据
		fmt.Printf("  📥 首次获取数据，请求 %d 根K线...\n", limit)
		newData, err := cf.fetcher.FetchOHLCV(symbol, interval, limit)
		if err != nil {
			return nil, err
		}
		
		// 保存到缓存
		cf.cache.Set(symbol, interval, newData)
		fmt.Printf("  💾 已缓存 %d 根K线数据\n", len(newData))
		
		return newData, nil
	}
	
	// 有缓存但数据不够，尝试增量更新
	return cf.fetchIncremental(symbol, interval, limit, cachedData)
}

// fetchIncremental 增量获取数据
func (cf *CachedFetcher) fetchIncremental(symbol string, interval string, limit int, cachedData []types.OHLCV) ([]types.OHLCV, error) {
	// 获取最新时间
	latestTime := cachedData[len(cachedData)-1].Time
	
	// 计算需要获取多少新数据
	// 根据时间间隔计算从最新时间到现在有多少根K线
	timeDiff := time.Since(latestTime)
	expectedNewBars := cf.calculateExpectedBars(interval, timeDiff)
	
	// 如果预期新数据很少，且缓存数据足够，直接使用缓存
	if expectedNewBars < 5 && len(cachedData) >= limit {
		fmt.Printf("  ⚡ 使用缓存数据（最新: %s）\n", latestTime.Format("01-02 15:04"))
		start := len(cachedData) - limit
		if start < 0 {
			start = 0
		}
		return cachedData[start:], nil
	}
	
	// 获取新数据（多获取一些以确保覆盖）
	fetchLimit := expectedNewBars + 10
	if fetchLimit < 50 {
		fetchLimit = 50 // 至少获取50根
	}
	
	fmt.Printf("  🔄 增量更新：获取最新 %d 根K线...\n", fetchLimit)
	newData, err := cf.fetcher.FetchOHLCV(symbol, interval, fetchLimit)
	if err != nil {
		// 如果获取失败，返回缓存数据
		fmt.Printf("  ⚠️  获取新数据失败，使用缓存数据\n")
		if len(cachedData) >= limit {
			start := len(cachedData) - limit
			return cachedData[start:], nil
		}
		return cachedData, nil
	}
	
	// 更新缓存
	cf.cache.Update(symbol, interval, newData)
	fmt.Printf("  ✅ 更新成功，新增 %d 根K线\n", len(newData))
	
	// 重新获取更新后的缓存
	updatedData, _ := cf.cache.Get(symbol, interval)
	if len(updatedData) >= limit {
		start := len(updatedData) - limit
		return updatedData[start:], nil
	}
	
	return updatedData, nil
}

// calculateExpectedBars 计算预期的K线数量
func (cf *CachedFetcher) calculateExpectedBars(interval string, duration time.Duration) int {
	switch interval {
	case "15m":
		return int(duration.Minutes() / 15)
	case "30m":
		return int(duration.Minutes() / 30)
	case "1h":
		return int(duration.Hours())
	case "4h":
		return int(duration.Hours() / 4)
	case "1d":
		return int(duration.Hours() / 24)
	default:
		return int(duration.Hours()) // 默认按小时
	}
}

// ClearCache 清除缓存
func (cf *CachedFetcher) ClearCache(symbol, interval string) {
	cf.cache.Clear(symbol, interval)
	fmt.Printf("  🗑️  已清除 %s %s 的缓存\n", symbol, interval)
}