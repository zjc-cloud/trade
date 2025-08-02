# B圈加密货币分析系统使用指南

## 系统概述

这是一个专门为加密货币市场设计的综合分析系统，特别适合B圈投资者使用。系统涵盖了主流币种（BTC、ETH、BNB）以及热门板块的深度分析。

## 核心功能

### 1. 多维度技术分析
- **传统技术指标**：MA、MACD、RSI、ADX、布林带等
- **币圈特有指标**：
  - 恐慌贪婪指数
  - 巨鲸活动监测
  - 资金费率分析
  - 市场吸筹/派发阶段判断

### 2. 多时间框架分析
- 15分钟、1小时、4小时、日线综合分析
- 自动识别趋势一致性
- 提供交易偏向和信心度评估

### 3. 板块轮动监测
- **预设板块**：
  - Layer1（ETH、SOL、AVAX等）
  - Layer2（ARB、OP、MATIC等）
  - DeFi蓝筹（UNI、AAVE、LINK等）
  - Meme币（DOGE、SHIB、PEPE等）
  - AI概念（FET、RNDR等）

### 4. 币种相关性分析
- BTC相关性评估
- 市场耦合度判断
- 最优持仓组合建议

## 快速使用

### 1. 基础分析
```bash
# 分析主流三大币种
python3 crypto_main.py --once --watchlist top3

# 分析特定币种
python3 crypto_main.py --symbols BTC/USDT ETH/USDT --once

# 使用不同时间框架
python3 crypto_main.py --symbols SOL/USDT --timeframe 1h --once
```

### 2. 持续监控
```bash
# 监控主流币种（每5分钟更新）
python3 crypto_main.py --watchlist top3 --interval 300

# 监控DeFi板块
python3 crypto_main.py --watchlist defi_blue --interval 600

# 监控Layer2板块
python3 crypto_main.py --watchlist layer2
```

### 3. 快速分析脚本
```bash
# 运行综合分析
python3 crypto_quick_analysis.py

# 监控模式
python3 crypto_quick_analysis.py --monitor

# 分析特定币种
python3 crypto_quick_analysis.py --coin BTC/USDT
```

## 预设监控列表

- **top3**: BTC/USDT, ETH/USDT, BNB/USDT
- **defi_blue**: UNI/USDT, AAVE/USDT, LINK/USDT, MKR/USDT
- **layer2**: ARB/USDT, OP/USDT, MATIC/USDT
- **trending**: SOL/USDT, AVAX/USDT, INJ/USDT, TIA/USDT
- **meme**: DOGE/USDT, SHIB/USDT, PEPE/USDT, BONK/USDT
- **ai_narrative**: FET/USDT, RNDR/USDT, AGIX/USDT, OCEAN/USDT

## 关键指标解读

### 恐慌贪婪指数
- 0-25：极度恐慌（买入机会）
- 25-45：恐慌（谨慎买入）
- 45-55：中性（观望）
- 55-75：贪婪（谨慎卖出）
- 75-100：极度贪婪（卖出机会）

### 巨鲸活动
- 监测大额交易（超过均值+2倍标准差）
- 区分巨鲸买入和卖出
- 判断市场主力动向

### 市场阶段
- **吸筹阶段**：价格下跌但OBV上升，主力建仓
- **派发阶段**：价格上涨但OBV下降，主力出货
- **趋势确认**：价格与成交量同向，趋势延续

### 多时间框架信号
- **完全对齐**：所有时间框架趋势一致，信号最强
- **偏多/偏空对齐**：大部分时间框架一致
- **趋势分歧**：不同时间框架信号矛盾，谨慎操作

## 交易策略建议

### 1. 根据恐慌贪婪指数
- **< 20**：分批建仓，越跌越买
- **20-40**：适量买入，保留资金
- **40-60**：高抛低吸，波段操作
- **60-80**：逐步减仓，锁定利润
- **> 80**：清仓观望，等待回调

### 2. 根据市场耦合度
- **高度耦合**：减少持仓币种，重仓龙头
- **中度耦合**：关注板块轮动，快进快出
- **低度耦合**：精选个币，分散投资

### 3. 根据巨鲸活动
- **巨鲸买入**：跟随建仓，但设止损
- **巨鲸卖出**：及时止盈，降低仓位
- **无巨鲸活动**：散户行情，技术分析为主

## 风险管理

1. **仓位管理**
   - BTC/ETH：最多50%仓位
   - 主流山寨：最多30%仓位
   - 小币种：最多10%仓位

2. **止损设置**
   - 短线：3-5%
   - 中线：5-10%
   - 长线：10-15%

3. **资金分配**
   - 交易资金：30%
   - 波段资金：40%
   - 长期持有：30%

## 注意事项

1. API限制：避免频繁请求，使用缓存功能
2. 数据延迟：实时数据可能有1-5分钟延迟
3. 市场异常：极端行情下技术指标可能失效
4. 投资风险：本系统仅供参考，不构成投资建议

## 常见问题

**Q: API被限制怎么办？**
A: 使用缓存模式或等待15分钟后重试

**Q: 如何添加新币种？**
A: 编辑crypto_config.py添加币种配置

**Q: 指标参数如何调整？**
A: 修改indicators目录下相应文件的默认参数

**Q: 如何接入其他交易所？**
A: 在data_fetcher.py中添加新的交易所类

祝你在B圈投资顺利！记住：DYOR（Do Your Own Research）！