package backtest

import (
	"fmt"
	"math"
	"time"

	"github.com/zjc/go-crypto-analyzer/pkg/analysis"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// PositionType 仓位类型
type PositionType int

const (
	NoPosition PositionType = iota
	LongPosition
	ShortPosition
)

// BacktesterV2 支持做空的回测器
type BacktesterV2 struct {
	analyzer          *analysis.TrendAnalyzer
	evidenceCollector *analysis.EvidenceCollector
	
	// 回测参数
	initialCapital float64
	feeRate        float64
	slippage       float64
	
	// 策略参数
	longThreshold   float64  // 做多阈值
	shortThreshold  float64  // 做空阈值
	closeThreshold  float64  // 平仓阈值
	stopLoss        float64  // 止损百分比
	takeProfit      float64  // 止盈百分比
	
	// 新增：双向交易
	allowShort      bool
	positionType    PositionType
	maxLeverage     float64  // 最大杠杆
	
	// 新增：改进策略
	improvedStrategy *ImprovedBidirectionalStrategy
	useImproved      bool
	currentStopLoss  float64  // 当前止损价
}

// TradeV2 交易记录（支持做空）
type TradeV2 struct {
	EntryTime    time.Time
	EntryPrice   float64
	EntrySignal  string
	ExitTime     time.Time
	ExitPrice    float64
	ExitSignal   string
	Direction    string     // "LONG" or "SHORT"
	Profit       float64
	ProfitPct    float64
	Size         float64
}

// BacktestResultV2 回测结果
type BacktestResultV2 struct {
	Symbol          string
	Period          string
	InitialCapital  float64
	FinalCapital    float64
	TotalReturn     float64
	TotalReturnPct  float64
	MaxDrawdown     float64
	MaxDrawdownPct  float64
	WinRate         float64
	TotalTrades     int
	LongTrades      int
	ShortTrades     int
	WinningTrades   int
	LosingTrades    int
	AverageWin      float64
	AverageLoss     float64
	ProfitFactor    float64
	SharpeRatio     float64
	CalmarRatio     float64
	Trades          []TradeV2
}

// NewBacktesterV2 创建支持做空的回测器
func NewBacktesterV2(initialCapital float64) *BacktesterV2 {
	return &BacktesterV2{
		analyzer:          analysis.NewTrendAnalyzer(),
		evidenceCollector: analysis.NewEvidenceCollector(),
		initialCapital:    initialCapital,
		feeRate:           0.001,
		slippage:          0.0005,
		longThreshold:     0.5,
		shortThreshold:    -0.5,
		closeThreshold:    0.0,
		stopLoss:          0.03,
		takeProfit:        0.06,
		allowShort:        true,
		positionType:      NoPosition,
		maxLeverage:       2.0,
		improvedStrategy:  NewImprovedBidirectionalStrategy(),
		useImproved:       false,
	}
}

// EnableShort 启用做空
func (bt *BacktesterV2) EnableShort(enable bool) {
	bt.allowShort = enable
}

// SetThresholds 设置阈值
func (bt *BacktesterV2) SetThresholds(long, short, close float64) {
	bt.longThreshold = long
	bt.shortThreshold = short
	bt.closeThreshold = close
}

// UseImprovedStrategy 使用改进的策略
func (bt *BacktesterV2) UseImprovedStrategy(use bool) {
	bt.useImproved = use
}

// RunBacktestV2 运行支持做空的回测
func (bt *BacktesterV2) RunBacktestV2(symbol string, data []types.OHLCV) (*BacktestResultV2, error) {
	if len(data) < 200 {
		return nil, fmt.Errorf("insufficient data for backtest")
	}
	
	result := &BacktestResultV2{
		Symbol:         symbol,
		Period:         fmt.Sprintf("%s to %s", data[0].Time.Format("2006-01-02"), data[len(data)-1].Time.Format("2006-01-02")),
		InitialCapital: bt.initialCapital,
		FinalCapital:   bt.initialCapital,
		Trades:         make([]TradeV2, 0),
	}
	
	// 状态变量
	capital := bt.initialCapital
	position := 0.0
	entryPrice := 0.0
	entryTime := time.Time{}
	entrySignal := ""
	maxCapital := capital
	bt.positionType = NoPosition
	
	// 滑动窗口分析
	for i := 100; i < len(data); i++ {
		window := data[i-100 : i+1]
		currentPrice := window[len(window)-1].Close
		currentTime := window[len(window)-1].Time
		
		// 执行技术分析
		analysisResult, err := bt.analyzer.AnalyzeComprehensive(window)
		if err != nil {
			continue
		}
		
		// 收集证据
		bt.evidenceCollector.Clear()
		bt.evidenceCollector.AnalyzeMAEvidence(analysisResult.MAAnalysis, currentPrice)
		bt.evidenceCollector.AnalyzeMACDEvidence(analysisResult.MACDAnalysis)
		bt.evidenceCollector.AnalyzeRSIEvidence(analysisResult.Momentum.RSI)
		bt.evidenceCollector.AnalyzeSREvidence(currentPrice, analysisResult.SupportResistance)
		
		// 计算价格变化率（用于成交量分析）
		priceChange := 0.0
		if i > 0 {
			priceChange = (currentPrice - data[i-1].Close) / data[i-1].Close
		}
		bt.evidenceCollector.AnalyzeVolumeEvidence(analysisResult.Volume, priceChange)
		
		// 获取信号强度
		summary := bt.evidenceCollector.GetSummary()
		totalStrength := summary["totalStrength"].(float64)
		
		// 检查止损止盈
		if bt.positionType != NoPosition && position > 0 {
			var profitPct float64
			
			if bt.positionType == LongPosition {
				profitPct = (currentPrice - entryPrice) / entryPrice
			} else { // ShortPosition
				profitPct = (entryPrice - currentPrice) / entryPrice
			}
			
			// 止损
			if profitPct <= -bt.stopLoss {
				exitPrice := currentPrice
				var profit float64
				if bt.positionType == LongPosition {
					exitPrice = currentPrice * (1 - bt.slippage - bt.feeRate)
					profit = position * (exitPrice - entryPrice)
					capital = position * exitPrice
				} else {
					exitPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
					profit = position * (entryPrice - exitPrice)
					capital = position * (2*entryPrice - exitPrice)
				}
				
				trade := TradeV2{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  "止损",
					Direction:   bt.getPositionString(),
					Profit:      profit,
					ProfitPct:   profitPct,
					Size:        position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				bt.positionType = NoPosition
				continue
			}
			
			// 止盈
			if profitPct >= bt.takeProfit {
				exitPrice := currentPrice
				var profit float64
				if bt.positionType == LongPosition {
					exitPrice = currentPrice * (1 - bt.slippage - bt.feeRate)
					profit = position * (exitPrice - entryPrice)
					capital = position * exitPrice
				} else {
					exitPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
					profit = position * (entryPrice - exitPrice)
					capital = position * (2*entryPrice - exitPrice)
				}
				
				trade := TradeV2{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  "止盈",
					Direction:   bt.getPositionString(),
					Profit:      profit,
					ProfitPct:   profitPct,
					Size:        position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				bt.positionType = NoPosition
				continue
			}
		}
		
		// 交易信号
		if bt.positionType == NoPosition {
			if bt.useImproved {
				// 使用改进策略
				marketRegime := bt.improvedStrategy.AnalyzeMarketRegime(analysisResult, window)
				
				// 做多信号
				if shouldLong, reason := bt.improvedStrategy.ShouldOpenLong(analysisResult, summary, marketRegime, window); shouldLong {
					entryPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
					position = capital / entryPrice
					capital = 0
					entryTime = currentTime
					entrySignal = reason
					bt.positionType = LongPosition
					
					// 计算动态止损
					atr := bt.calculateATR(window, 14)
					bt.currentStopLoss = bt.improvedStrategy.GetDynamicStopLoss(entryPrice, currentPrice, LongPosition, atr)
					
				// 做空信号
				} else if bt.allowShort {
					if shouldShort, reason := bt.improvedStrategy.ShouldOpenShort(analysisResult, summary, marketRegime, window); shouldShort {
						entryPrice = currentPrice * (1 - bt.slippage - bt.feeRate)
						position = capital / entryPrice
						capital = 0
						entryTime = currentTime
						entrySignal = reason
						bt.positionType = ShortPosition
						
						// 计算动态止损
						atr := bt.calculateATR(window, 14)
						bt.currentStopLoss = bt.improvedStrategy.GetDynamicStopLoss(entryPrice, currentPrice, ShortPosition, atr)
					}
				}
			} else {
				// 使用原始策略
				// 做多信号
				if totalStrength > bt.longThreshold {
					entryPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
					position = capital / entryPrice
					capital = 0
					entryTime = currentTime
					entrySignal = fmt.Sprintf("做多(强度:%.2f)", totalStrength)
					bt.positionType = LongPosition
					
				// 做空信号
				} else if bt.allowShort && totalStrength < bt.shortThreshold {
					entryPrice = currentPrice * (1 - bt.slippage - bt.feeRate)
					position = capital / entryPrice
					capital = 0
					entryTime = currentTime
					entrySignal = fmt.Sprintf("做空(强度:%.2f)", totalStrength)
					bt.positionType = ShortPosition
				}
			}
			
		} else if bt.positionType == LongPosition {
			// 多头平仓信号
			shouldExit := false
			exitReason := ""
			
			if bt.useImproved {
				// 使用改进策略的出场逻辑
				marketRegime := bt.improvedStrategy.AnalyzeMarketRegime(analysisResult, window)
				shouldExit, exitReason = bt.improvedStrategy.ShouldCloseLong(analysisResult, summary, entryPrice, currentPrice, marketRegime)
				
				// 更新动态止损
				if !shouldExit && bt.improvedStrategy.dynamicStopLoss {
					atr := bt.calculateATR(window, 14)
					newStopLoss := bt.improvedStrategy.GetDynamicStopLoss(entryPrice, currentPrice, LongPosition, atr)
					if newStopLoss > bt.currentStopLoss {
						bt.currentStopLoss = newStopLoss
					}
					
					// 检查动态止损
					if currentPrice <= bt.currentStopLoss {
						shouldExit = true
						exitReason = fmt.Sprintf("动态止损(%.2f)", bt.currentStopLoss)
					}
				}
			} else {
				// 原始策略逻辑
				if totalStrength < bt.closeThreshold {
					shouldExit = true
					exitReason = fmt.Sprintf("平多(强度:%.2f)", totalStrength)
				}
			}
			
			if shouldExit {
				exitPrice := currentPrice * (1 - bt.slippage - bt.feeRate)
				profit := position * (exitPrice - entryPrice)
				capital = position * exitPrice
				
				trade := TradeV2{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  exitReason,
					Direction:   "LONG",
					Profit:      profit,
					ProfitPct:   profit / (position * entryPrice),
					Size:        position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				bt.positionType = NoPosition
				
				// 立即检查是否可以反向开仓
				if bt.allowShort && totalStrength < bt.shortThreshold {
					entryPrice = currentPrice * (1 - bt.slippage - bt.feeRate)
					position = capital / entryPrice
					capital = 0
					entryTime = currentTime
					entrySignal = fmt.Sprintf("反手做空(强度:%.2f)", totalStrength)
					bt.positionType = ShortPosition
				}
			}
			
		} else if bt.positionType == ShortPosition {
			// 空头平仓信号
			shouldExit := false
			exitReason := ""
			
			if bt.useImproved {
				// 使用改进策略的出场逻辑
				marketRegime := bt.improvedStrategy.AnalyzeMarketRegime(analysisResult, window)
				shouldExit, exitReason = bt.improvedStrategy.ShouldCloseShort(analysisResult, summary, entryPrice, currentPrice, marketRegime)
				
				// 更新动态止损
				if !shouldExit && bt.improvedStrategy.dynamicStopLoss {
					atr := bt.calculateATR(window, 14)
					newStopLoss := bt.improvedStrategy.GetDynamicStopLoss(entryPrice, currentPrice, ShortPosition, atr)
					if newStopLoss < bt.currentStopLoss {
						bt.currentStopLoss = newStopLoss
					}
					
					// 检查动态止损
					if currentPrice >= bt.currentStopLoss {
						shouldExit = true
						exitReason = fmt.Sprintf("动态止损(%.2f)", bt.currentStopLoss)
					}
				}
			} else {
				// 原始策略逻辑
				if totalStrength > -bt.closeThreshold {
					shouldExit = true
					exitReason = fmt.Sprintf("平空(强度:%.2f)", totalStrength)
				}
			}
			
			if shouldExit {
				exitPrice := currentPrice * (1 + bt.slippage + bt.feeRate)
				profit := position * (entryPrice - exitPrice)
				capital = position * (2*entryPrice - exitPrice)
				
				trade := TradeV2{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  exitReason,
					Direction:   "SHORT",
					Profit:      profit,
					ProfitPct:   profit / (position * entryPrice),
					Size:        position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				bt.positionType = NoPosition
				
				// 立即检查是否可以反向开仓
				if totalStrength > bt.longThreshold {
					entryPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
					position = capital / entryPrice
					capital = 0
					entryTime = currentTime
					entrySignal = fmt.Sprintf("反手做多(强度:%.2f)", totalStrength)
					bt.positionType = LongPosition
				}
			}
		}
		
		// 更新最高资金
		currentCapital := capital
		if position > 0 {
			if bt.positionType == LongPosition {
				currentCapital = position * currentPrice
			} else {
				currentCapital = position * (2*entryPrice - currentPrice)
			}
		}
		if currentCapital > maxCapital {
			maxCapital = currentCapital
		}
		
		// 计算回撤
		drawdown := (maxCapital - currentCapital) / maxCapital
		if drawdown > result.MaxDrawdownPct {
			result.MaxDrawdownPct = drawdown
			result.MaxDrawdown = maxCapital - currentCapital
		}
	}
	
	// 强制平仓未平仓位
	if position > 0 {
		exitPrice := data[len(data)-1].Close
		profit := 0.0
		
		if bt.positionType == LongPosition {
			exitPrice = exitPrice * (1 - bt.slippage - bt.feeRate)
			profit = position * (exitPrice - entryPrice)
			capital = position * exitPrice
		} else {
			exitPrice = exitPrice * (1 + bt.slippage + bt.feeRate)
			profit = position * (entryPrice - exitPrice)
			capital = position * (2*entryPrice - exitPrice)
		}
		
		trade := TradeV2{
			EntryTime:   entryTime,
			EntryPrice:  entryPrice,
			EntrySignal: entrySignal,
			ExitTime:    data[len(data)-1].Time,
			ExitPrice:   exitPrice,
			ExitSignal:  "回测结束平仓",
			Direction:   bt.getPositionString(),
			Profit:      profit,
			ProfitPct:   profit / (position * entryPrice),
			Size:        position,
		}
		result.Trades = append(result.Trades, trade)
	}
	
	// 计算统计数据
	result.FinalCapital = capital
	result.TotalReturn = capital - bt.initialCapital
	result.TotalReturnPct = result.TotalReturn / bt.initialCapital
	result.TotalTrades = len(result.Trades)
	
	totalWin := 0.0
	totalLoss := 0.0
	returns := make([]float64, 0)
	
	for _, trade := range result.Trades {
		returns = append(returns, trade.ProfitPct)
		
		if trade.Direction == "LONG" {
			result.LongTrades++
		} else {
			result.ShortTrades++
		}
		
		if trade.Profit > 0 {
			result.WinningTrades++
			totalWin += trade.Profit
		} else {
			result.LosingTrades++
			totalLoss += math.Abs(trade.Profit)
		}
	}
	
	if result.TotalTrades > 0 {
		result.WinRate = float64(result.WinningTrades) / float64(result.TotalTrades)
	}
	
	if result.WinningTrades > 0 {
		result.AverageWin = totalWin / float64(result.WinningTrades)
	}
	
	if result.LosingTrades > 0 {
		result.AverageLoss = totalLoss / float64(result.LosingTrades)
	}
	
	if totalLoss > 0 {
		result.ProfitFactor = totalWin / totalLoss
	}
	
	// 计算夏普比率
	if len(returns) > 0 {
		avgReturn := 0.0
		for _, r := range returns {
			avgReturn += r
		}
		avgReturn /= float64(len(returns))
		
		variance := 0.0
		for _, r := range returns {
			variance += math.Pow(r-avgReturn, 2)
		}
		
		if len(returns) > 1 {
			variance /= float64(len(returns) - 1)
			stdDev := math.Sqrt(variance)
			if stdDev > 0 {
				result.SharpeRatio = avgReturn / stdDev * math.Sqrt(8760)
			}
		}
	}
	
	// 计算卡尔玛比率
	if result.MaxDrawdownPct > 0 {
		annualizedReturn := result.TotalReturnPct * 365 / float64(len(data)/24)
		result.CalmarRatio = annualizedReturn / result.MaxDrawdownPct
	}
	
	return result, nil
}

// getPositionString 获取仓位字符串
func (bt *BacktesterV2) getPositionString() string {
	switch bt.positionType {
	case LongPosition:
		return "LONG"
	case ShortPosition:
		return "SHORT"
	default:
		return "NONE"
	}
}

// calculateATR 计算ATR（平均真实波幅）
func (bt *BacktesterV2) calculateATR(data []types.OHLCV, period int) float64 {
	if len(data) < period+1 {
		return 0
	}
	
	// 计算真实波幅
	trueRanges := make([]float64, 0)
	for i := len(data) - period; i < len(data); i++ {
		if i == 0 {
			continue
		}
		
		// TR = max(H-L, abs(H-PC), abs(L-PC))
		highLow := data[i].High - data[i].Low
		highPrevClose := math.Abs(data[i].High - data[i-1].Close)
		lowPrevClose := math.Abs(data[i].Low - data[i-1].Close)
		
		tr := math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
		trueRanges = append(trueRanges, tr)
	}
	
	// 计算ATR
	if len(trueRanges) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, tr := range trueRanges {
		sum += tr
	}
	
	return sum / float64(len(trueRanges))
}