package backtest

import (
	"fmt"
	"math"
	"time"

	"github.com/zjc/go-crypto-analyzer/pkg/analysis"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// BacktestResult 回测结果
type BacktestResult struct {
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
	WinningTrades   int
	LosingTrades    int
	AverageWin      float64
	AverageLoss     float64
	ProfitFactor    float64
	SharpeRatio     float64
	Trades          []Trade
}

// Trade 交易记录
type Trade struct {
	EntryTime   time.Time
	EntryPrice  float64
	EntrySignal string
	ExitTime    time.Time
	ExitPrice   float64
	ExitSignal  string
	Profit      float64
	ProfitPct   float64
	Holding     float64 // 持仓数量
}

// Backtester 回测器
type Backtester struct {
	analyzer          *analysis.TrendAnalyzer
	evidenceCollector *analysis.EvidenceCollector
	
	// 回测参数
	initialCapital float64
	feeRate        float64  // 手续费率
	slippage       float64  // 滑点
	
	// 策略参数
	entryThreshold  float64  // 入场阈值
	exitThreshold   float64  // 出场阈值
	stopLoss        float64  // 止损百分比
	takeProfit      float64  // 止盈百分比
	
	// 新增：策略接口
	strategy        TradingStrategy
	useStrategy     bool
}

// NewBacktester 创建回测器
func NewBacktester(initialCapital float64) *Backtester {
	return &Backtester{
		analyzer:          analysis.NewTrendAnalyzer(),
		evidenceCollector: analysis.NewEvidenceCollector(),
		initialCapital:    initialCapital,
		feeRate:           0.001, // 0.1% 手续费
		slippage:          0.0005, // 0.05% 滑点
		entryThreshold:    0.5,    // 综合强度>0.5做多，<-0.5做空
		exitThreshold:     -0.2,   // 反向信号平仓
		stopLoss:          0.05,   // 5%止损
		takeProfit:        0.10,   // 10%止盈
	}
}

// RunBacktest 运行回测
func (bt *Backtester) RunBacktest(symbol string, data []types.OHLCV) (*BacktestResult, error) {
	if len(data) < 200 {
		return nil, fmt.Errorf("insufficient data for backtest (need at least 200 candles)")
	}
	
	result := &BacktestResult{
		Symbol:         symbol,
		Period:         fmt.Sprintf("%s to %s", data[0].Time.Format("2006-01-02"), data[len(data)-1].Time.Format("2006-01-02")),
		InitialCapital: bt.initialCapital,
		FinalCapital:   bt.initialCapital,
		Trades:         make([]Trade, 0),
	}
	
	// 状态变量
	capital := bt.initialCapital
	position := 0.0          // 当前持仓
	entryPrice := 0.0        // 入场价格
	entryTime := time.Time{} // 入场时间
	entrySignal := ""        // 入场信号
	maxCapital := capital    // 最高资金
	
	// 滑动窗口分析
	for i := 100; i < len(data); i++ {
		// 使用前100个数据点进行技术分析
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
		
		// 获取信号强度
		summary := bt.evidenceCollector.GetSummary()
		totalStrength := summary["totalStrength"].(float64)
		
		// 检查止损止盈
		if position > 0 {
			profitPct := (currentPrice - entryPrice) / entryPrice
			
			// 止损
			if profitPct <= -bt.stopLoss {
				exitPrice := currentPrice * (1 - bt.slippage - bt.feeRate)
				profit := position * (exitPrice - entryPrice)
				capital += position * exitPrice
				
				trade := Trade{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  "止损",
					Profit:      profit,
					ProfitPct:   profit / (position * entryPrice),
					Holding:     position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				continue
			}
			
			// 止盈
			if profitPct >= bt.takeProfit {
				exitPrice := currentPrice * (1 - bt.slippage - bt.feeRate)
				profit := position * (exitPrice - entryPrice)
				capital += position * exitPrice
				
				trade := Trade{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  "止盈",
					Profit:      profit,
					ProfitPct:   profit / (position * entryPrice),
					Holding:     position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
				continue
			}
		}
		
		// 交易信号
		if bt.useStrategy && bt.strategy != nil {
			// 使用策略接口
			if shouldEnter, reason := bt.strategy.ShouldEnter(analysisResult, summary, position); shouldEnter {
				entryPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
				position = capital / entryPrice
				capital = 0
				entryTime = currentTime
				entrySignal = reason
				
				// 更新止损止盈
				bt.stopLoss = (entryPrice - bt.strategy.GetStopLoss(entryPrice, analysisResult)) / entryPrice
				bt.takeProfit = (bt.strategy.GetTakeProfit(entryPrice, analysisResult) - entryPrice) / entryPrice
			} else if position > 0 {
				if shouldExit, reason := bt.strategy.ShouldExit(analysisResult, summary, position, entryPrice); shouldExit {
					exitPrice := currentPrice * (1 - bt.slippage - bt.feeRate)
					profit := position * (exitPrice - entryPrice)
					capital = position * exitPrice
					
					trade := Trade{
						EntryTime:   entryTime,
						EntryPrice:  entryPrice,
						EntrySignal: entrySignal,
						ExitTime:    currentTime,
						ExitPrice:   exitPrice,
						ExitSignal:  reason,
						Profit:      profit,
						ProfitPct:   profit / (position * entryPrice),
						Holding:     position,
					}
					result.Trades = append(result.Trades, trade)
					
					position = 0.0
				}
			}
		} else {
			// 使用原始逻辑
			if position == 0 && totalStrength > bt.entryThreshold {
				// 做多信号
				entryPrice = currentPrice * (1 + bt.slippage + bt.feeRate)
				position = capital / entryPrice
				capital = 0
				entryTime = currentTime
				entrySignal = fmt.Sprintf("做多(强度:%.2f)", totalStrength)
				
			} else if position > 0 && totalStrength < bt.exitThreshold {
				// 平仓信号
				exitPrice := currentPrice * (1 - bt.slippage - bt.feeRate)
				profit := position * (exitPrice - entryPrice)
				capital = position * exitPrice
				
				trade := Trade{
					EntryTime:   entryTime,
					EntryPrice:  entryPrice,
					EntrySignal: entrySignal,
					ExitTime:    currentTime,
					ExitPrice:   exitPrice,
					ExitSignal:  fmt.Sprintf("平仓(强度:%.2f)", totalStrength),
					Profit:      profit,
					ProfitPct:   profit / (position * entryPrice),
					Holding:     position,
				}
				result.Trades = append(result.Trades, trade)
				
				position = 0.0
			}
		}
		
		// 更新最高资金（用于计算最大回撤）
		currentCapital := capital
		if position > 0 {
			currentCapital = position * currentPrice
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
	
	// 如果还有持仓，按最后价格平仓
	if position > 0 {
		exitPrice := data[len(data)-1].Close * (1 - bt.slippage - bt.feeRate)
		profit := position * (exitPrice - entryPrice)
		capital = position * exitPrice
		
		trade := Trade{
			EntryTime:   entryTime,
			EntryPrice:  entryPrice,
			EntrySignal: entrySignal,
			ExitTime:    data[len(data)-1].Time,
			ExitPrice:   exitPrice,
			ExitSignal:  "回测结束平仓",
			Profit:      profit,
			ProfitPct:   profit / (position * entryPrice),
			Holding:     position,
		}
		result.Trades = append(result.Trades, trade)
	}
	
	// 计算统计指标
	result.FinalCapital = capital
	result.TotalReturn = capital - bt.initialCapital
	result.TotalReturnPct = result.TotalReturn / bt.initialCapital
	result.TotalTrades = len(result.Trades)
	
	totalWin := 0.0
	totalLoss := 0.0
	returns := make([]float64, 0)
	
	for _, trade := range result.Trades {
		returns = append(returns, trade.ProfitPct)
		
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
				// 年化夏普比率（假设1小时K线，一年8760小时）
				result.SharpeRatio = avgReturn / stdDev * math.Sqrt(8760)
			}
		}
	}
	
	return result, nil
}

// SetStrategy 设置策略参数
func (bt *Backtester) SetStrategy(entryThreshold, exitThreshold, stopLoss, takeProfit float64) {
	bt.entryThreshold = entryThreshold
	bt.exitThreshold = exitThreshold
	bt.stopLoss = stopLoss
	bt.takeProfit = takeProfit
}

// SetFees 设置费用参数
func (bt *Backtester) SetFees(feeRate, slippage float64) {
	bt.feeRate = feeRate
	bt.slippage = slippage
}

// SetTradingStrategy 设置交易策略
func (bt *Backtester) SetTradingStrategy(strategy TradingStrategy) {
	bt.strategy = strategy
	bt.useStrategy = true
}