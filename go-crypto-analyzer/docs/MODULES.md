# 模块使用指南

## 项目模块化设计

本项目分为两个独立但相互关联的模块：

### 模块1：市场分析（Market Analysis）
实时获取和分析市场数据，提供当前市场趋势判断。

### 模块2：交易回测（Trading & Backtesting）  
基于历史数据验证交易策略，评估策略效果。

## 模块1：市场分析使用指南

### 功能说明
- 获取实时价格和K线数据
- 计算技术指标（MA、MACD、RSI、ADX等）
- 分析市场趋势和交易信号
- 提供买卖建议和风险提示

### 快速使用

```bash
# 分析BTC当前市场状况
./crypto-analyzer

# 分析ETH，使用4小时K线
./crypto-analyzer -s ETHUSDT -i 4h

# 开启持续监控模式
./crypto-analyzer -c -d 300
```

### 输出解读

```
技术指标：
- RSI < 30: 超卖，可能反弹
- RSI > 70: 超买，可能回调
- MACD金叉: 看涨信号
- ADX > 25: 趋势明显

市场趋势：
- 强劲上涨: 多个指标看涨，建议做多
- 震荡: 方向不明，建议观望
- 强劲下跌: 多个指标看跌，建议做空或离场
```

## 模块2：交易回测使用指南

### 功能说明
- 支持多种交易策略回测
- 双向交易（做多/做空）
- 详细的收益和风险分析
- 交易记录和统计报告

### 策略选择

1. **初学者**：使用Simple策略
   ```bash
   ./backtest -s BTCUSDT -d 30
   ```

2. **趋势交易者**：使用Trend策略
   ```bash
   ./backtest -s BTCUSDT -d 30 --strategy trend
   ```

3. **短线交易者**：使用Momentum策略
   ```bash
   ./backtest -s BTCUSDT -d 30 --strategy momentum
   ```

4. **双向交易**：使用backtest-v2
   ```bash
   ./backtest-v2 -s BTCUSDT -d 30 --improved
   ```

### 参数优化建议

根据市场状况调整参数：

**牛市参数**：
- 入场阈值：0.3（更容易入场）
- 止损：5%（给予更多空间）
- 止盈：15%（追求更高收益）

**熊市参数**：
- 入场阈值：0.8（更严格筛选）
- 止损：3%（快速止损）
- 止盈：5%（见好就收）

**震荡市参数**：
- 使用均值回归策略
- 入场阈值：0.6
- 止损：2%
- 止盈：3%

## 模块间协作

### 工作流程

1. **市场分析**：先用模块1分析当前市场
   ```bash
   ./crypto-analyzer -s BTCUSDT
   ```

2. **策略选择**：根据市场状态选择合适策略
   - 趋势明显 → Trend策略
   - 震荡市场 → Reversal策略
   - 高波动 → Momentum策略

3. **历史验证**：用模块2回测验证
   ```bash
   ./backtest -s BTCUSDT -d 60 --strategy trend
   ```

4. **参数调优**：根据回测结果优化参数
   ```bash
   ./backtest-v2 -s BTCUSDT -d 60 --long 0.6 --short -0.6
   ```

### 实战示例

**场景1：发现上涨趋势**
```bash
# 1. 分析当前市场
./crypto-analyzer -s SOLUSDT
# 输出：强劲上涨趋势，ADX=45

# 2. 回测趋势策略
./backtest -s SOLUSDT -d 30 --strategy trend
# 结果：收益15%，胜率60%

# 3. 优化参数后再测
./backtest-v2 -s SOLUSDT -d 30 --long 0.4 --improved
```

**场景2：震荡市场操作**
```bash
# 1. 分析发现震荡
./crypto-analyzer -s ETHUSDT
# 输出：横盘震荡，ADX=15

# 2. 使用均值回归策略
./backtest -s ETHUSDT -d 30 --strategy reversal
```

## 常见问题

### Q: 为什么回测总是亏损？
A: 可能原因：
- 交易太频繁，手续费累积
- 止损设置太紧，频繁止损
- 策略不适合当前市场
- 参数需要优化

### Q: 如何提高胜率？
A: 建议：
- 提高入场门槛（如0.8以上）
- 只在趋势明确时交易
- 使用更长的时间周期（4h、1d）
- 结合多个确认信号

### Q: 实盘能用吗？
A: 本项目仅供学习研究，实盘需要：
- 更完善的风险控制
- 考虑滑点和真实成交
- 小资金测试验证
- 严格的资金管理

## 进阶技巧

### 组合使用多个时间框架
```bash
# 大周期看趋势
./crypto-analyzer -s BTCUSDT -i 1d

# 小周期找入场点
./crypto-analyzer -s BTCUSDT -i 1h
```

### 批量回测多个币种
```bash
for coin in BTCUSDT ETHUSDT BNBUSDT; do
    echo "Testing $coin..."
    ./backtest -s $coin -d 30 --strategy trend
done
```

### 导出分析结果
```bash
# 导出到文件
./crypto-analyzer -s BTCUSDT -o btc_analysis.json

# 定期记录
./crypto-analyzer -c -d 3600 >> market_log.txt
```