# 快速开始指南

## 5分钟上手

### 1. 安装
```bash
git clone https://github.com/zjc/go-crypto-analyzer.git
cd go-crypto-analyzer
go mod download
```

### 2. 构建
```bash
# 使用脚本构建所有模块
./run.sh
# 选择 4 构建项目
```

或手动构建：
```bash
go build -o crypto-analyzer ./cmd/crypto-analyzer
go build -o backtest ./cmd/backtest
go build -o backtest-v2 ./cmd/backtest-v2
```

### 3. 使用

#### 方式一：使用交互式菜单
```bash
./run.sh
```

#### 方式二：直接运行

**查看BTC当前市场：**
```bash
./crypto-analyzer
```

**回测交易策略：**
```bash
./backtest -s BTCUSDT -d 30
```

## 典型使用流程

### 步骤1：分析市场
```bash
# 查看BTC趋势
./crypto-analyzer -s BTCUSDT

# 如果显示"强劲上涨趋势"，继续步骤2
```

### 步骤2：选择策略
- 上涨趋势 → 使用 trend 策略
- 震荡市场 → 使用 reversal 策略
- 高波动 → 使用 momentum 策略

### 步骤3：回测验证
```bash
# 测试趋势策略
./backtest -s BTCUSDT -d 30 --strategy trend
```

### 步骤4：优化参数
如果结果不理想，调整参数：
```bash
# 提高入场门槛，减少交易次数
./backtest-v2 -s BTCUSDT -d 30 --long 0.8 --short -0.8
```

## 重要提示

### 关于回测亏损
回测亏损很正常，原因包括：
- 手续费累积（每次0.15%）
- 市场噪音导致频繁止损
- 策略参数需要优化

### 改善建议
1. **减少交易频率**：提高入场阈值到0.8以上
2. **放宽止损**：从3%调整到5%
3. **选对市场**：趋势明显时效果更好
4. **延长周期**：使用4h或1d K线

## 常用命令速查

```bash
# 市场分析
./crypto-analyzer                    # 分析BTC
./crypto-analyzer -s ETHUSDT        # 分析ETH
./crypto-analyzer -c -d 300         # 持续监控，5分钟更新

# 策略回测
./backtest -s BTCUSDT -d 30         # 基础回测
./backtest -s BTCUSDT -d 30 --strategy trend  # 趋势策略

# 双向交易
./backtest-v2 -s BTCUSDT -d 30      # 支持做空
./backtest-v2 -s BTCUSDT -d 30 --improved  # 改进策略
```

## 下一步

1. 阅读 [模块使用指南](docs/MODULES.md) 了解详细功能
2. 查看 README.md 了解完整参数说明
3. 根据自己的交易风格调整策略参数