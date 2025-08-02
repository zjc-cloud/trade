package backtest

import (
	"fmt"
	"math"

	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// ImprovedBidirectionalStrategy 改进的双向交易策略
type ImprovedBidirectionalStrategy struct {
	// 市场状态检测
	trendStrengthThreshold float64  // ADX阈值
	volatilityPeriod       int      // 波动率计算周期
	
	// 入场条件
	longSignalThreshold    float64  // 做多信号阈值
	shortSignalThreshold   float64  // 做空信号阈值
	volumeConfirmation     float64  // 成交量确认倍数
	
	// 风险管理
	dynamicStopLoss        bool     // 是否使用动态止损
	atrMultiplier          float64  // ATR止损倍数
	trailingStop           bool     // 是否使用移动止损
	
	// 市场状态
	currentMarketRegime    string   // trending/ranging/volatile
	positionBias           string   // long/short/neutral
}

// NewImprovedBidirectionalStrategy 创建改进的双向策略
func NewImprovedBidirectionalStrategy() *ImprovedBidirectionalStrategy {
	return &ImprovedBidirectionalStrategy{
		trendStrengthThreshold: 25.0,
		volatilityPeriod:       20,
		longSignalThreshold:    0.6,
		shortSignalThreshold:   -0.6,
		volumeConfirmation:     1.5,
		dynamicStopLoss:        true,
		atrMultiplier:          2.0,
		trailingStop:           true,
		currentMarketRegime:    "unknown",
		positionBias:           "neutral",
	}
}

// AnalyzeMarketRegime 分析市场状态
func (s *ImprovedBidirectionalStrategy) AnalyzeMarketRegime(analysis *types.Analysis, data []types.OHLCV) string {
	adx := analysis.TrendStrength.ADX
	
	// 计算最近的波动率
	volatility := s.calculateVolatility(data, s.volatilityPeriod)
	avgVolatility := s.calculateVolatility(data, 50)
	
	// 趋势强度分析
	if adx > 40 {
		if analysis.MAAnalysis.Trend == types.StrongUptrend {
			s.positionBias = "long"
			return "strong_uptrend"
		} else if analysis.MAAnalysis.Trend == types.StrongDowntrend {
			s.positionBias = "short"
			return "strong_downtrend"
		}
	}
	
	// 区间震荡市场
	if adx < 20 && volatility < avgVolatility*0.8 {
		s.positionBias = "neutral"
		return "ranging"
	}
	
	// 高波动市场
	if volatility > avgVolatility*1.5 {
		return "volatile"
	}
	
	// 普通趋势市场
	if adx > s.trendStrengthThreshold {
		if analysis.MAAnalysis.Trend == types.Uptrend {
			s.positionBias = "long"
			return "uptrend"
		} else if analysis.MAAnalysis.Trend == types.Downtrend {
			s.positionBias = "short"
			return "downtrend"
		}
	}
	
	return "neutral"
}

// calculateVolatility 计算波动率
func (s *ImprovedBidirectionalStrategy) calculateVolatility(data []types.OHLCV, period int) float64 {
	if len(data) < period {
		return 0
	}
	
	returns := make([]float64, period-1)
	for i := len(data)-period+1; i < len(data); i++ {
		returns[i-(len(data)-period+1)] = math.Log(data[i].Close / data[i-1].Close)
	}
	
	// 计算标准差
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))
	
	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(returns))
	
	return math.Sqrt(variance) * math.Sqrt(252*24) // 年化波动率（小时数据）
}

// ShouldOpenLong 判断是否开多
func (s *ImprovedBidirectionalStrategy) ShouldOpenLong(
	analysis *types.Analysis, 
	evidenceSummary map[string]interface{},
	marketRegime string,
	data []types.OHLCV,
) (bool, string) {
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	
	// 市场状态过滤
	switch marketRegime {
	case "strong_downtrend", "downtrend":
		// 下跌趋势中不做多
		return false, ""
	case "ranging":
		// 区间震荡需要更强的信号
		if totalStrength < s.longSignalThreshold*1.2 {
			return false, ""
		}
	case "volatile":
		// 高波动市场谨慎做多
		if analysis.Momentum.RSI > 60 {
			return false, ""
		}
	}
	
	// 基本信号强度检查
	if totalStrength < s.longSignalThreshold {
		return false, ""
	}
	
	// 成交量确认
	if analysis.Volume.VolumeRatio < s.volumeConfirmation {
		return false, ""
	}
	
	// 技术指标确认
	confirmations := 0
	
	// MACD确认
	if analysis.MACDAnalysis.MACD > analysis.MACDAnalysis.Signal && 
	   analysis.MACDAnalysis.Histogram > 0 {
		confirmations++
	}
	
	// RSI确认（不能超买）
	if analysis.Momentum.RSI > 30 && analysis.Momentum.RSI < 70 {
		confirmations++
	}
	
	// 价格位置确认
	if analysis.CurrentPrice > analysis.MAAnalysis.MA5 && 
	   analysis.CurrentPrice > analysis.MAAnalysis.MA10 {
		confirmations++
	}
	
	// 布林带确认
	bb := s.calculateBollingerBands(data, 20, 2)
	if analysis.CurrentPrice > bb.lower && analysis.CurrentPrice < bb.middle {
		confirmations++
	}
	
	// 需要至少3个确认信号
	if confirmations < 3 {
		return false, ""
	}
	
	reason := fmt.Sprintf("做多信号(强度:%.2f,确认:%d,市场:%s)", 
		totalStrength, confirmations, marketRegime)
	
	return true, reason
}

// ShouldOpenShort 判断是否开空
func (s *ImprovedBidirectionalStrategy) ShouldOpenShort(
	analysis *types.Analysis,
	evidenceSummary map[string]interface{},
	marketRegime string,
	data []types.OHLCV,
) (bool, string) {
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	
	// 市场状态过滤
	switch marketRegime {
	case "strong_uptrend", "uptrend":
		// 上涨趋势中不做空
		return false, ""
	case "ranging":
		// 区间震荡需要更强的信号
		if totalStrength > s.shortSignalThreshold*1.2 {
			return false, ""
		}
	case "volatile":
		// 高波动市场谨慎做空
		if analysis.Momentum.RSI < 40 {
			return false, ""
		}
	}
	
	// 基本信号强度检查
	if totalStrength > s.shortSignalThreshold {
		return false, ""
	}
	
	// 成交量确认
	if analysis.Volume.VolumeRatio < s.volumeConfirmation {
		return false, ""
	}
	
	// 技术指标确认
	confirmations := 0
	
	// MACD确认
	if analysis.MACDAnalysis.MACD < analysis.MACDAnalysis.Signal && 
	   analysis.MACDAnalysis.Histogram < 0 {
		confirmations++
	}
	
	// RSI确认（不能超卖）
	if analysis.Momentum.RSI < 70 && analysis.Momentum.RSI > 30 {
		confirmations++
	}
	
	// 价格位置确认
	if analysis.CurrentPrice < analysis.MAAnalysis.MA5 && 
	   analysis.CurrentPrice < analysis.MAAnalysis.MA10 {
		confirmations++
	}
	
	// 布林带确认
	bb := s.calculateBollingerBands(data, 20, 2)
	if analysis.CurrentPrice < bb.upper && analysis.CurrentPrice > bb.middle {
		confirmations++
	}
	
	// 需要至少3个确认信号
	if confirmations < 3 {
		return false, ""
	}
	
	reason := fmt.Sprintf("做空信号(强度:%.2f,确认:%d,市场:%s)", 
		totalStrength, confirmations, marketRegime)
	
	return true, reason
}

// ShouldCloseLong 判断是否平多
func (s *ImprovedBidirectionalStrategy) ShouldCloseLong(
	analysis *types.Analysis,
	evidenceSummary map[string]interface{},
	entryPrice float64,
	currentPrice float64,
	marketRegime string,
) (bool, string) {
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	profitPct := (currentPrice - entryPrice) / entryPrice
	
	// 止盈条件
	if profitPct > 0.05 && totalStrength < 0 {
		return true, fmt.Sprintf("止盈平多(收益:%.2f%%)", profitPct*100)
	}
	
	// 趋势反转
	if marketRegime == "downtrend" || marketRegime == "strong_downtrend" {
		return true, "趋势反转平多"
	}
	
	// 技术指标背离
	if analysis.MACDAnalysis.Histogram < 0 && analysis.Momentum.RSI > 70 {
		return true, "技术背离平多"
	}
	
	// 跌破关键支撑
	if currentPrice < analysis.MAAnalysis.MA20*0.98 {
		return true, "跌破MA20平多"
	}
	
	// 强烈看跌信号
	if totalStrength < -0.8 {
		return true, fmt.Sprintf("强烈看跌平多(强度:%.2f)", totalStrength)
	}
	
	return false, ""
}

// ShouldCloseShort 判断是否平空
func (s *ImprovedBidirectionalStrategy) ShouldCloseShort(
	analysis *types.Analysis,
	evidenceSummary map[string]interface{},
	entryPrice float64,
	currentPrice float64,
	marketRegime string,
) (bool, string) {
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	profitPct := (entryPrice - currentPrice) / entryPrice
	
	// 止盈条件
	if profitPct > 0.05 && totalStrength > 0 {
		return true, fmt.Sprintf("止盈平空(收益:%.2f%%)", profitPct*100)
	}
	
	// 趋势反转
	if marketRegime == "uptrend" || marketRegime == "strong_uptrend" {
		return true, "趋势反转平空"
	}
	
	// 技术指标背离
	if analysis.MACDAnalysis.Histogram > 0 && analysis.Momentum.RSI < 30 {
		return true, "技术背离平空"
	}
	
	// 突破关键阻力
	if currentPrice > analysis.MAAnalysis.MA20*1.02 {
		return true, "突破MA20平空"
	}
	
	// 强烈看涨信号
	if totalStrength > 0.8 {
		return true, fmt.Sprintf("强烈看涨平空(强度:%.2f)", totalStrength)
	}
	
	return false, ""
}

// GetDynamicStopLoss 获取动态止损价格
func (s *ImprovedBidirectionalStrategy) GetDynamicStopLoss(
	entryPrice float64,
	currentPrice float64,
	positionType PositionType,
	atr float64,
) float64 {
	
	if !s.dynamicStopLoss {
		// 固定止损
		if positionType == LongPosition {
			return entryPrice * 0.97
		} else {
			return entryPrice * 1.03
		}
	}
	
	// ATR动态止损
	stopDistance := atr * s.atrMultiplier
	
	if positionType == LongPosition {
		stopLoss := currentPrice - stopDistance
		// 移动止损：只能向上移动
		if s.trailingStop && currentPrice > entryPrice*1.02 {
			return math.Max(stopLoss, entryPrice*1.005) // 保证至少保本
		}
		return math.Max(stopLoss, entryPrice*0.95) // 最大损失5%
	} else {
		stopLoss := currentPrice + stopDistance
		// 移动止损：只能向下移动
		if s.trailingStop && currentPrice < entryPrice*0.98 {
			return math.Min(stopLoss, entryPrice*0.995) // 保证至少保本
		}
		return math.Min(stopLoss, entryPrice*1.05) // 最大损失5%
	}
}

// BollingerBands 布林带
type BollingerBands struct {
	upper  float64
	middle float64
	lower  float64
}

// calculateBollingerBands 计算布林带
func (s *ImprovedBidirectionalStrategy) calculateBollingerBands(data []types.OHLCV, period int, stdDev float64) BollingerBands {
	if len(data) < period {
		return BollingerBands{}
	}
	
	// 计算SMA
	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i].Close
	}
	sma := sum / float64(period)
	
	// 计算标准差
	variance := 0.0
	for i := len(data) - period; i < len(data); i++ {
		variance += math.Pow(data[i].Close-sma, 2)
	}
	std := math.Sqrt(variance / float64(period))
	
	return BollingerBands{
		upper:  sma + std*stdDev,
		middle: sma,
		lower:  sma - std*stdDev,
	}
}