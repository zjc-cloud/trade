# API限制问题解决方案

## 当前问题
- Binance API: 418错误，IP被临时封禁
- Yahoo Finance: 429错误，请求过多

## 立即解决方案

### 1. 等待解封
- Binance: 通常5-60分钟自动解封
- Yahoo: 通常1小时后重置

### 2. 使用代理或VPN
```bash
# 通过代理访问
export HTTP_PROXY=http://your-proxy:port
export HTTPS_PROXY=http://your-proxy:port
./crypto-analyzer
```

### 3. 减少数据请求量
```bash
# 减少K线数量
./crypto-analyzer -l 50  # 只获取50根K线

# 使用更大的时间间隔
./crypto-analyzer -i 4h  # 使用4小时K线，请求量更少
```

## 长期解决方案

### 1. 实现请求缓存
在代码中添加缓存机制，避免重复请求相同数据：

```go
// pkg/data/cache.go
type DataCache struct {
    cache map[string]CachedData
    mu    sync.RWMutex
}

type CachedData struct {
    Data      []types.OHLCV
    Timestamp time.Time
}

func (c *DataCache) Get(key string) ([]types.OHLCV, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    cached, exists := c.cache[key]
    if !exists {
        return nil, false
    }
    
    // 缓存5分钟有效
    if time.Since(cached.Timestamp) > 5*time.Minute {
        return nil, false
    }
    
    return cached.Data, true
}
```

### 2. 实现请求限流
添加限流器控制请求频率：

```go
// pkg/data/ratelimit.go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter() *RateLimiter {
    // 每秒最多2个请求
    return &RateLimiter{
        limiter: rate.NewLimiter(2, 1),
    }
}

func (r *RateLimiter) Wait(ctx context.Context) error {
    return r.limiter.Wait(ctx)
}
```

### 3. 使用多个数据源
实现数据源切换机制：

```go
// pkg/data/multi_source.go
type MultiSourceFetcher struct {
    sources []Fetcher
    current int
}

func (m *MultiSourceFetcher) FetchOHLCV(symbol string, interval string, limit int) ([]types.OHLCV, error) {
    for i := 0; i < len(m.sources); i++ {
        source := m.sources[m.current]
        data, err := source.FetchOHLCV(symbol, interval, limit)
        if err == nil {
            return data, nil
        }
        
        // 切换到下一个数据源
        m.current = (m.current + 1) % len(m.sources)
    }
    
    return nil, fmt.Errorf("all data sources failed")
}
```

### 4. 本地数据存储
保存历史数据到本地：

```go
// pkg/data/local_storage.go
func SaveToFile(symbol string, data []types.OHLCV) error {
    filename := fmt.Sprintf("data/%s_%s.json", symbol, time.Now().Format("20060102"))
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, jsonData, 0644)
}

func LoadFromFile(symbol string) ([]types.OHLCV, error) {
    pattern := fmt.Sprintf("data/%s_*.json", symbol)
    files, err := filepath.Glob(pattern)
    if err != nil || len(files) == 0 {
        return nil, fmt.Errorf("no local data found")
    }
    
    // 使用最新的文件
    sort.Strings(files)
    latestFile := files[len(files)-1]
    
    data, err := os.ReadFile(latestFile)
    if err != nil {
        return nil, err
    }
    
    var ohlcv []types.OHLCV
    return ohlcv, json.Unmarshal(data, &ohlcv)
}
```

## 配置文件优化

创建 `.env` 文件管理API配置：

```env
# Binance API配置
BINANCE_API_WEIGHT_LIMIT=1200
BINANCE_REQUEST_INTERVAL=500ms

# Yahoo Finance配置  
YAHOO_REQUEST_LIMIT=100
YAHOO_REQUEST_INTERVAL=1s

# 缓存配置
CACHE_ENABLED=true
CACHE_TTL=5m

# 代理配置（可选）
HTTP_PROXY=
HTTPS_PROXY=
```

## 使用其他数据源

### 1. CoinGecko API
```bash
# 免费tier: 10-50 calls/minute
https://api.coingecko.com/api/v3/coins/bitcoin/ohlc?vs_currency=usd&days=1
```

### 2. CryptoCompare API
```bash
# 免费tier: 100,000 calls/month
https://min-api.cryptocompare.com/data/v2/histohour?fsym=BTC&tsym=USD&limit=100
```

### 3. Alpha Vantage
```bash
# 免费tier: 5 API requests/minute
https://www.alphavantage.co/query?function=DIGITAL_CURRENCY_DAILY&symbol=BTC&market=USD
```

## 监控和预警

添加API使用情况监控：

```go
type APIMonitor struct {
    requestCount int64
    errorCount   int64
    lastReset    time.Time
}

func (m *APIMonitor) RecordRequest() {
    atomic.AddInt64(&m.requestCount, 1)
}

func (m *APIMonitor) RecordError() {
    atomic.AddInt64(&m.errorCount, 1)
}

func (m *APIMonitor) ShouldSlowDown() bool {
    // 如果错误率超过10%，应该减慢请求
    if m.errorCount > 0 && float64(m.errorCount)/float64(m.requestCount) > 0.1 {
        return true
    }
    return false
}
```

## 最佳实践

1. **批量请求**：一次获取更多数据，减少请求次数
2. **错误重试**：实现指数退避重试机制
3. **健康检查**：定期检查API状态
4. **降级策略**：API不可用时使用缓存或本地数据
5. **监控告警**：记录API使用情况，提前预警

通过这些方案，可以有效避免API限制问题，提高系统稳定性。