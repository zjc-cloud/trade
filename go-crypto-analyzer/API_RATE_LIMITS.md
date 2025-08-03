# API速率限制说明

## 回答您的问题

**Q: 为什么使用雅虎的数据首次就失败了？请求100根K线请求了几次，一次还是100次？**

**A: 只请求了1次，不是100次。**

### 详细解释

1. **请求机制**：
   - Yahoo Finance API使用单次HTTP请求获取所有数据
   - 请求100根K线 = 1个HTTP请求
   - URL格式：`https://query1.finance.yahoo.com/v8/finance/chart/BTC-USD?period1=xxx&period2=xxx&interval=1h`

2. **失败原因**：
   - 错误代码429：Too Many Requests（请求过多）
   - 即使是第一次请求也可能失败
   - 原因：您的IP地址可能在之前已经触发了限制

## API限制详情

### Binance API
- **错误代码**: 418 (I'm a teapot) 
- **含义**: IP被临时封禁
- **限制规则**:
  - 每分钟1200权重
  - 获取K线数据：权重1-2
  - 封禁时长：5-60分钟

### Yahoo Finance API
- **错误代码**: 429 (Too Many Requests)
- **含义**: 超过速率限制
- **限制规则**:
  - 无官方文档说明具体限制
  - 经验值：每小时约100-200请求
  - 重置时间：通常1小时

## 为什么首次请求就失败？

1. **IP级别限制**：
   - API限制是基于IP地址的
   - 即使是新程序，如果IP已被标记，首次请求也会失败

2. **可能的触发原因**：
   - 之前运行过其他程序访问相同API
   - 短时间内多次重启程序
   - 共享IP（如公司网络、VPN）

3. **时间窗口**：
   - 限制通常基于滑动窗口
   - 例如：过去1小时内的总请求数

## 缓存如何帮助？

我们实现的缓存机制可以：

1. **减少请求频率**：
   ```
   首次运行：1次API请求 → 缓存100根K线
   后续运行：0次API请求（使用缓存）
   更新数据：1次API请求（仅获取新增部分）
   ```

2. **智能更新策略**：
   - 检查缓存时效性
   - 计算需要的新数据量
   - 只请求增量数据

## 建议解决方案

### 立即措施
1. **等待重置**：
   - Binance: 等待5-60分钟
   - Yahoo: 等待1小时

2. **使用缓存**：
   ```bash
   # 清除旧缓存
   ./crypto-analyzer --clear-cache
   
   # 减少数据量，使用更长的缓存时间
   ./crypto-analyzer -l 30 --cache-ttl 30
   ```

3. **切换网络**：
   - 使用手机热点
   - 使用VPN（注意：某些API可能封禁VPN）

### 长期方案
1. **使用多个数据源轮换**
2. **实现请求队列和延迟**
3. **部署到云服务器（新IP）**
4. **申请API密钥（更高限额）**

## 监控建议

在代码中添加请求计数器：
```go
var requestCount int
var lastReset time.Time

func trackRequest() {
    if time.Since(lastReset) > time.Hour {
        requestCount = 0
        lastReset = time.Now()
    }
    requestCount++
    fmt.Printf("API请求数: %d/小时\n", requestCount)
}
```