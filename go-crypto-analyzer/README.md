# Go Crypto Analyzer

一个用于加密货币市场分析和交易策略回测的Go语言项目。

## 项目结构

项目分为两个主要模块：

### 模块1：市场分析模块（Market Analysis）
用于获取和分析当前市场数据，提供趋势判断和交易信号。

### 模块2：交易回测模块（Trading & Backtesting）
基于历史数据测试交易策略，评估策略表现。

## 功能特性

- 🔍 **实时市场分析**：获取并分析加密货币市场数据
- 📊 **技术指标计算**：MA、MACD、RSI、ADX等多种指标
- 📈 **趋势判断**：基于多指标交叉验证的趋势分析
- 🔄 **双向交易**：支持做多和做空策略
- 📉 **策略回测**：历史数据回测，评估策略效果
- 🎯 **风险管理**：动态止损、止盈设置
- 🎨 **美观的CLI界面**: 彩色输出，ASCII图表
- ⚡ **实时监控**: 支持持续监控模式

## 快速开始

### 环境要求
- Go 1.20+
- 网络连接（访问Binance API）

### 安装

```bash
git clone https://github.com/zjc/go-crypto-analyzer.git
cd go-crypto-analyzer
go mod download
```

## 模块1：市场分析

### 构建
```bash
go build -o crypto-analyzer ./cmd/crypto-analyzer
```

### 运行

基本用法：
```bash
./crypto-analyzer
```

指定交易对分析：
```bash
./crypto-analyzer -s ETHUSDT
```

使用Yahoo Finance数据源：
```bash
./crypto-analyzer -s BTC-USD --yahoo
```

持续监控模式：
```bash
./crypto-analyzer -c -d 300  # 每5分钟更新
```

### 参数说明
- `-s, --symbol`: 交易对符号（默认：BTCUSDT）
- `-i, --interval`: K线时间间隔（默认：1h）
- `-f, --fear-greed`: 显示恐慌贪婪指数
- `-y, --yahoo`: 使用Yahoo Finance数据源
- `-o, --output`: 导出分析结果到文件
- `-c, --continuous`: 持续监控模式
- `-d, --delay`: 监控间隔秒数（默认：300）

### 输出示例
```
================================================================================
📊 加密货币市场分析 - 2025-08-03 01:00:00
================================================================================

💰 当前价格: $112,930.24
📈 24h涨跌: -2.15%
💎 市场趋势: 震荡

技术指标：
┌─────────────┬──────────┬──────────┬────────┐
│ 指标        │ 数值     │ 参考值   │ 状态   │
├─────────────┼──────────┼──────────┼────────┤
│ RSI(14)     │ 45.2     │ 超买>70  │ 中性   │
│ MACD        │ -234.5   │ 信号线   │ 看跌   │
│ ADX         │ 18.5     │ 趋势>25  │ 震荡   │
└─────────────┴──────────┴──────────┴────────┘

综合判断：
✅ 市场处于震荡状态，建议观望
```

## 模块2：交易回测

### 构建
```bash
# 基础回测器
go build -o backtest ./cmd/backtest

# 双向交易回测器
go build -o backtest-v2 ./cmd/backtest-v2
```

### 运行

#### 基础回测（仅做多）
```bash
# 使用简单策略
./backtest -s BTCUSDT -d 30

# 使用趋势跟踪策略
./backtest -s BTCUSDT -d 30 --strategy trend

# 使用动量策略
./backtest -s BTCUSDT -d 30 --strategy momentum
```

#### 双向交易回测
```bash
# 基础双向策略
./backtest-v2 -s BTCUSDT -d 30

# 使用改进策略（动态止损）
./backtest-v2 -s BTCUSDT -d 30 --improved

# 禁用做空
./backtest-v2 -s BTCUSDT -d 30 --enable-short=false
```

### 回测参数
- `-s, --symbol`: 交易对（默认：BTCUSDT）
- `-d, --days`: 回测天数（默认：30）
- `-c, --capital`: 初始资金（默认：10000）
- `-S, --strategy`: 策略类型 (simple|trend|momentum|reversal|combo)
- `--improved`: 使用改进的自适应策略
- `--enable-short`: 启用做空（默认：true）

### 策略说明

1. **Simple（简单阈值）**：基于信号强度的简单策略
2. **Trend（趋势跟踪）**：跟随市场趋势，使用ADX过滤
3. **Momentum（动量突破）**：捕捉动量爆发机会
4. **Reversal（均值回归）**：在超卖区域买入
5. **Combo（组合策略）**：根据市场状态自动切换策略

## 模块间关系

模块2（交易回测）复用了模块1的核心组件：
- `pkg/indicators`: 技术指标计算
- `pkg/analysis`: 趋势分析和证据收集
- `pkg/data`: 数据获取接口
- `pkg/types`: 公共数据类型

```
go-crypto-analyzer/
├── cmd/
│   ├── crypto-analyzer/    # 模块1: 市场分析
│   ├── backtest/          # 模块2: 基础回测
│   └── backtest-v2/       # 模块2: 双向交易回测
├── pkg/                   # 共享核心组件
│   ├── indicators/        # 技术指标计算
│   ├── data/             # 数据获取模块
│   ├── analysis/         # 趋势分析和证据收集
│   ├── backtest/         # 回测引擎
│   └── types/            # 类型定义
└── internal/             # 内部模块
    ├── config/           # 配置管理
    └── alert/            # 预警系统
```

## 回测结果分析

### 为什么会亏损？

1. **交易成本**：每次交易0.15%的成本（手续费+滑点），频繁交易导致成本累积
2. **止损设置**：3%的止损在波动市场中太紧，经常被触发
3. **信号质量**：阈值设置过低，导致很多弱信号也触发交易
4. **市场状态**：策略未能适应不同市场状态（趋势/震荡/高波动）

### 改进建议

1. **提高入场门槛**：增加信号强度阈值，减少交易次数
2. **动态止损**：根据ATR调整止损距离
3. **市场过滤**：只在趋势明确时交易
4. **仓位管理**：根据信号强度调整仓位大小

## 配置文件

创建 `config.json` 自定义参数：
```json
{
  "analysis": {
    "default_symbol": "BTCUSDT",
    "default_interval": "1h"
  },
  "backtest": {
    "initial_capital": 10000,
    "fee_rate": 0.001,
    "slippage": 0.0005,
    "stop_loss": 0.05,
    "take_profit": 0.10
  }
}
```

## 常见问题

1. **API限制**：如遇到429错误，请稍后重试或使用Yahoo数据源
2. **数据不足**：回测需要至少200根K线数据
3. **内存占用**：长时间回测可能占用较多内存

## 开发计划

- [ ] 添加更多技术指标
- [ ] 支持多币种组合策略
- [ ] 实时交易接口
- [ ] Web界面
- [ ] 机器学习策略

## 许可证

MIT License