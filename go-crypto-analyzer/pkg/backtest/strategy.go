package backtest

import (
	"fmt"
	
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// StrategyType 策略类型
type StrategyType string

const (
	SimpleStrategy    StrategyType = "simple"    // 简单阈值策略
	TrendStrategy     StrategyType = "trend"     // 趋势跟踪策略
	MomentumStrategy  StrategyType = "momentum"  // 动量策略
	ReversalStrategy  StrategyType = "reversal"  // 反转策略
	ComboStrategy     StrategyType = "combo"     // 组合策略
)

// TradingStrategy 交易策略接口
type TradingStrategy interface {
	// ShouldEnter 判断是否应该入场
	ShouldEnter(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64) (bool, string)
	
	// ShouldExit 判断是否应该出场
	ShouldExit(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64, entryPrice float64) (bool, string)
	
	// GetStopLoss 获取止损价格
	GetStopLoss(entryPrice float64, analysis *types.Analysis) float64
	
	// GetTakeProfit 获取止盈价格
	GetTakeProfit(entryPrice float64, analysis *types.Analysis) float64
}

// TrendFollowingStrategy 趋势跟踪策略
type TrendFollowingStrategy struct {
	minADX          float64  // 最小ADX值
	minVolumeRatio  float64  // 最小成交量比
	entryThreshold  float64  // 入场阈值
	exitThreshold   float64  // 出场阈值
	useATRStop      bool     // 使用ATR止损
	atrMultiplier   float64  // ATR乘数
}

// NewTrendFollowingStrategy 创建趋势跟踪策略
func NewTrendFollowingStrategy() *TrendFollowingStrategy {
	return &TrendFollowingStrategy{
		minADX:         25.0,
		minVolumeRatio: 1.2,
		entryThreshold: 0.8,
		exitThreshold:  -0.3,
		useATRStop:     true,
		atrMultiplier:  2.0,
	}
}

// ShouldEnter 趋势策略入场条件
func (s *TrendFollowingStrategy) ShouldEnter(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64) (bool, string) {
	if position > 0 {
		return false, ""
	}
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	
	// 基本条件检查
	if totalStrength <= s.entryThreshold {
		return false, ""
	}
	
	// ADX过滤 - 只在趋势市场交易
	if analysis.TrendStrength.ADX < s.minADX {
		return false, ""
	}
	
	// 成交量确认
	if analysis.Volume.VolumeRatio < s.minVolumeRatio {
		return false, ""
	}
	
	// 价格位置检查 - 必须在中期均线上方
	if analysis.CurrentPrice < analysis.MAAnalysis.MA20 {
		return false, ""
	}
	
	// RSI过滤 - 避免追高
	if analysis.Momentum.RSI > 75 {
		return false, ""
	}
	
	// MACD确认
	if analysis.MACDAnalysis.Trend != "看涨" {
		return false, ""
	}
	
	reason := fmt.Sprintf("趋势买入(ADX:%.1f,强度:%.2f)", 
		analysis.TrendStrength.ADX, totalStrength)
	
	return true, reason
}

// ShouldExit 趋势策略出场条件
func (s *TrendFollowingStrategy) ShouldExit(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64, entryPrice float64) (bool, string) {
	if position <= 0 {
		return false, ""
	}
	
	totalStrength := evidenceSummary["totalStrength"].(float64)
	
	// 趋势反转信号
	if totalStrength < s.exitThreshold {
		return true, fmt.Sprintf("趋势反转(强度:%.2f)", totalStrength)
	}
	
	// 跌破关键均线
	if analysis.CurrentPrice < analysis.MAAnalysis.MA20 {
		return true, "跌破MA20"
	}
	
	// MACD死叉
	if analysis.MACDAnalysis.Trend == "看跌" && analysis.MACDAnalysis.Histogram < 0 {
		return true, "MACD死叉"
	}
	
	// 成交量异常
	if analysis.Volume.VolumeRatio > 3 && analysis.CurrentPrice < entryPrice {
		return true, "放量下跌"
	}
	
	return false, ""
}

// GetStopLoss 计算止损价
func (s *TrendFollowingStrategy) GetStopLoss(entryPrice float64, analysis *types.Analysis) float64 {
	// 使用MA20作为止损参考
	stopLoss := analysis.MAAnalysis.MA20
	
	// 但不能超过5%
	maxLoss := entryPrice * 0.95
	if stopLoss < maxLoss {
		stopLoss = maxLoss
	}
	
	return stopLoss
}

// GetTakeProfit 计算止盈价
func (s *TrendFollowingStrategy) GetTakeProfit(entryPrice float64, analysis *types.Analysis) float64 {
	// 使用阻力位作为止盈目标
	r1 := analysis.SupportResistance.Resistance["R1"]
	
	// 但至少要有5%的利润
	minProfit := entryPrice * 1.05
	if r1 < minProfit {
		return minProfit
	}
	
	return r1
}

// MomentumBreakoutStrategy 动量突破策略
type MomentumBreakoutStrategy struct {
	rsiThreshold    float64
	volumeThreshold float64
	breakoutPeriod  int
}

// NewMomentumBreakoutStrategy 创建动量突破策略
func NewMomentumBreakoutStrategy() *MomentumBreakoutStrategy {
	return &MomentumBreakoutStrategy{
		rsiThreshold:    60.0,
		volumeThreshold: 2.0,
		breakoutPeriod:  20,
	}
}

// ShouldEnter 动量策略入场
func (s *MomentumBreakoutStrategy) ShouldEnter(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64) (bool, string) {
	if position > 0 {
		return false, ""
	}
	
	// RSI动量确认
	if analysis.Momentum.RSI < s.rsiThreshold || analysis.Momentum.RSI > 80 {
		return false, ""
	}
	
	// 成交量突破
	if analysis.Volume.VolumeRatio < s.volumeThreshold {
		return false, ""
	}
	
	// MACD柱状图必须为正且增长
	if analysis.MACDAnalysis.Histogram <= 0 {
		return false, ""
	}
	
	// 价格必须突破所有短期均线
	if analysis.CurrentPrice <= analysis.MAAnalysis.MA5 ||
	   analysis.CurrentPrice <= analysis.MAAnalysis.MA10 ||
	   analysis.CurrentPrice <= analysis.MAAnalysis.MA20 {
		return false, ""
	}
	
	reason := fmt.Sprintf("动量突破(RSI:%.1f,Vol:%.1fx)", 
		analysis.Momentum.RSI, analysis.Volume.VolumeRatio)
	
	return true, reason
}

// ShouldExit 动量策略出场
func (s *MomentumBreakoutStrategy) ShouldExit(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64, entryPrice float64) (bool, string) {
	if position <= 0 {
		return false, ""
	}
	
	// RSI超买
	if analysis.Momentum.RSI > 80 {
		return true, "RSI超买"
	}
	
	// 动量衰竭
	if analysis.Momentum.RSI < 50 {
		return true, "动量衰竭"
	}
	
	// MACD柱状图转负
	if analysis.MACDAnalysis.Histogram < 0 {
		return true, "MACD转负"
	}
	
	// 跌破MA5
	if analysis.CurrentPrice < analysis.MAAnalysis.MA5 {
		return true, "跌破MA5"
	}
	
	return false, ""
}

// GetStopLoss 动量策略止损
func (s *MomentumBreakoutStrategy) GetStopLoss(entryPrice float64, analysis *types.Analysis) float64 {
	// 使用MA5作为动态止损
	stopLoss := analysis.MAAnalysis.MA5
	
	// 但不能超过3%
	maxLoss := entryPrice * 0.97
	if stopLoss < maxLoss {
		stopLoss = maxLoss
	}
	
	return stopLoss
}

// GetTakeProfit 动量策略止盈
func (s *MomentumBreakoutStrategy) GetTakeProfit(entryPrice float64, analysis *types.Analysis) float64 {
	// 动量策略使用较小的止盈目标
	return entryPrice * 1.06
}

// MeanReversionStrategy 均值回归策略
type MeanReversionStrategy struct {
	oversoldRSI     float64
	overboughtRSI   float64
	bollingerPeriod int
	bollingerStdDev float64
}

// NewMeanReversionStrategy 创建均值回归策略
func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		oversoldRSI:     30.0,
		overboughtRSI:   70.0,
		bollingerPeriod: 20,
		bollingerStdDev: 2.0,
	}
}

// ShouldEnter 均值回归入场
func (s *MeanReversionStrategy) ShouldEnter(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64) (bool, string) {
	if position > 0 {
		return false, ""
	}
	
	// RSI超卖
	if analysis.Momentum.RSI >= s.oversoldRSI {
		return false, ""
	}
	
	// 价格必须远离均线（超卖）
	deviation := (analysis.CurrentPrice - analysis.MAAnalysis.MA20) / analysis.MAAnalysis.MA20
	if deviation > -0.03 { // 必须低于MA20至少3%
		return false, ""
	}
	
	// ADX低于25，表示没有强趋势
	if analysis.TrendStrength.ADX > 25 {
		return false, ""
	}
	
	// 价格接近支撑位
	s1 := analysis.SupportResistance.Support["S1"]
	if analysis.CurrentPrice > s1*1.01 { // 必须接近S1（1%以内）
		return false, ""
	}
	
	reason := fmt.Sprintf("超卖反弹(RSI:%.1f,偏离:%.1f%%)", 
		analysis.Momentum.RSI, deviation*100)
	
	return true, reason
}

// ShouldExit 均值回归出场
func (s *MeanReversionStrategy) ShouldExit(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64, entryPrice float64) (bool, string) {
	if position <= 0 {
		return false, ""
	}
	
	// 回归均值
	if analysis.CurrentPrice >= analysis.MAAnalysis.MA20 {
		return true, "回归MA20"
	}
	
	// RSI恢复正常
	if analysis.Momentum.RSI > 50 {
		return true, "RSI恢复"
	}
	
	// 达到阻力位
	if analysis.CurrentPrice >= analysis.SupportResistance.Resistance["R1"]*0.99 {
		return true, "接近阻力"
	}
	
	// 止盈3%
	if analysis.CurrentPrice >= entryPrice*1.03 {
		return true, "达到止盈"
	}
	
	return false, ""
}

// GetStopLoss 均值回归止损
func (s *MeanReversionStrategy) GetStopLoss(entryPrice float64, analysis *types.Analysis) float64 {
	// 使用S2作为止损
	return analysis.SupportResistance.Support["S2"]
}

// GetTakeProfit 均值回归止盈
func (s *MeanReversionStrategy) GetTakeProfit(entryPrice float64, analysis *types.Analysis) float64 {
	// 目标是回到MA20
	target := analysis.MAAnalysis.MA20
	
	// 但至少要有3%利润
	minProfit := entryPrice * 1.03
	if target < minProfit {
		return minProfit
	}
	
	return target
}

// ComboAdaptiveStrategy 自适应组合策略
type ComboAdaptiveStrategy struct {
	trendStrategy     *TrendFollowingStrategy
	momentumStrategy  *MomentumBreakoutStrategy
	reversionStrategy *MeanReversionStrategy
	currentMode       string
}

// NewComboAdaptiveStrategy 创建自适应组合策略
func NewComboAdaptiveStrategy() *ComboAdaptiveStrategy {
	return &ComboAdaptiveStrategy{
		trendStrategy:     NewTrendFollowingStrategy(),
		momentumStrategy:  NewMomentumBreakoutStrategy(),
		reversionStrategy: NewMeanReversionStrategy(),
		currentMode:       "detecting",
	}
}

// DetectMarketCondition 检测市场状态
func (s *ComboAdaptiveStrategy) DetectMarketCondition(analysis *types.Analysis) string {
	adx := analysis.TrendStrength.ADX
	rsi := analysis.Momentum.RSI
	
	// 强趋势市场
	if adx > 35 {
		if analysis.MAAnalysis.Trend == types.Uptrend || analysis.MAAnalysis.Trend == types.StrongUptrend {
			return "trending"
		}
	}
	
	// 动量市场
	if adx > 20 && adx <= 35 && rsi > 50 && rsi < 70 {
		return "momentum"
	}
	
	// 超卖反弹机会
	if adx < 25 && rsi < 30 {
		return "reversion"
	}
	
	// 默认观望
	return "neutral"
}

// ShouldEnter 自适应策略入场
func (s *ComboAdaptiveStrategy) ShouldEnter(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64) (bool, string) {
	if position > 0 {
		return false, ""
	}
	
	// 检测市场状态
	marketCondition := s.DetectMarketCondition(analysis)
	s.currentMode = marketCondition
	
	switch marketCondition {
	case "trending":
		if enter, reason := s.trendStrategy.ShouldEnter(analysis, evidenceSummary, position); enter {
			return true, "[趋势模式] " + reason
		}
	case "momentum":
		if enter, reason := s.momentumStrategy.ShouldEnter(analysis, evidenceSummary, position); enter {
			return true, "[动量模式] " + reason
		}
	case "reversion":
		if enter, reason := s.reversionStrategy.ShouldEnter(analysis, evidenceSummary, position); enter {
			return true, "[反转模式] " + reason
		}
	}
	
	return false, ""
}

// ShouldExit 自适应策略出场
func (s *ComboAdaptiveStrategy) ShouldExit(analysis *types.Analysis, evidenceSummary map[string]interface{}, position float64, entryPrice float64) (bool, string) {
	if position <= 0 {
		return false, ""
	}
	
	// 根据入场模式选择出场策略
	switch s.currentMode {
	case "trending":
		return s.trendStrategy.ShouldExit(analysis, evidenceSummary, position, entryPrice)
	case "momentum":
		return s.momentumStrategy.ShouldExit(analysis, evidenceSummary, position, entryPrice)
	case "reversion":
		return s.reversionStrategy.ShouldExit(analysis, evidenceSummary, position, entryPrice)
	}
	
	// 默认止损
	if analysis.CurrentPrice < entryPrice*0.95 {
		return true, "默认止损5%"
	}
	
	return false, ""
}

// GetStopLoss 自适应策略止损
func (s *ComboAdaptiveStrategy) GetStopLoss(entryPrice float64, analysis *types.Analysis) float64 {
	switch s.currentMode {
	case "trending":
		return s.trendStrategy.GetStopLoss(entryPrice, analysis)
	case "momentum":
		return s.momentumStrategy.GetStopLoss(entryPrice, analysis)
	case "reversion":
		return s.reversionStrategy.GetStopLoss(entryPrice, analysis)
	}
	
	return entryPrice * 0.95
}

// GetTakeProfit 自适应策略止盈
func (s *ComboAdaptiveStrategy) GetTakeProfit(entryPrice float64, analysis *types.Analysis) float64 {
	switch s.currentMode {
	case "trending":
		return s.trendStrategy.GetTakeProfit(entryPrice, analysis)
	case "momentum":
		return s.momentumStrategy.GetTakeProfit(entryPrice, analysis)
	case "reversion":
		return s.reversionStrategy.GetTakeProfit(entryPrice, analysis)
	}
	
	return entryPrice * 1.05
}