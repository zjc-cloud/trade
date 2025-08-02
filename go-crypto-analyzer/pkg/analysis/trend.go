package analysis

import (
	"fmt"

	"github.com/zjc/go-crypto-analyzer/pkg/indicators"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// TrendAnalyzer analyzes market trends
type TrendAnalyzer struct {
	indicators *indicators.TechnicalIndicators
}

// NewTrendAnalyzer creates a new TrendAnalyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{
		indicators: indicators.NewTechnicalIndicators(),
	}
}

// AnalyzeComprehensive performs comprehensive analysis on OHLCV data
func (ta *TrendAnalyzer) AnalyzeComprehensive(data []types.OHLCV) (*types.Analysis, error) {
	if len(data) < 50 {
		return nil, fmt.Errorf("insufficient data for analysis (need at least 50 candles)")
	}

	// Extract price data
	closes := extractCloses(data)
	highs := extractHighs(data)
	lows := extractLows(data)
	volumes := extractVolumes(data)

	// Moving Average Analysis
	maAnalysis := ta.analyzeMovingAverages(closes)

	// MACD Analysis
	macdAnalysis := ta.indicators.MACD(closes, 12, 26, 9)

	// Momentum Analysis
	rsi := ta.indicators.RSI(closes, 14)
	momentumAnalysis := ta.analyzeMomentum(rsi)

	// Trend Strength Analysis
	adx := ta.indicators.ADX(highs, lows, closes, 14)
	trendStrength := ta.analyzeTrendStrength(adx)

	// Volume Analysis
	volumeAnalysis := ta.indicators.VolumeAnalysis(volumes, 20)

	// Support and Resistance
	lastCandle := data[len(data)-1]
	srAnalysis := ta.indicators.PivotPoints(lastCandle.High, lastCandle.Low, lastCandle.Close)

	// Overall trend determination
	overallTrend, trendScore := ta.determineOverallTrend(maAnalysis, macdAnalysis, momentumAnalysis)

	return &types.Analysis{
		Symbol:            "", // Will be set by caller
		CurrentPrice:      closes[len(closes)-1],
		Timestamp:         data[len(data)-1].Time,
		OverallTrend:      overallTrend,
		TrendScore:        trendScore,
		MAAnalysis:        maAnalysis,
		MACDAnalysis:      macdAnalysis,
		Momentum:          momentumAnalysis,
		TrendStrength:     trendStrength,
		Volume:            volumeAnalysis,
		SupportResistance: srAnalysis,
	}, nil
}

// analyzeMovingAverages analyzes moving average trends
func (ta *TrendAnalyzer) analyzeMovingAverages(closes []float64) types.MAAnalysis {
	ma5 := ta.indicators.SMA(closes, 5)
	ma10 := ta.indicators.SMA(closes, 10)
	ma20 := ta.indicators.SMA(closes, 20)
	ma50 := ta.indicators.SMA(closes, 50)
	ma200 := ta.indicators.SMA(closes, 200)

	currentPrice := closes[len(closes)-1]
	
	// Get latest MA values
	lastMA5 := getLastValue(ma5, 5)
	lastMA10 := getLastValue(ma10, 10)
	lastMA20 := getLastValue(ma20, 20)
	lastMA50 := getLastValue(ma50, 50)
	lastMA200 := getLastValue(ma200, 200)

	// Calculate MA signals
	signals := 0.0
	if currentPrice > lastMA5 {
		signals += 1
	} else {
		signals -= 1
	}

	if lastMA5 > lastMA10 {
		signals += 1
	} else {
		signals -= 1
	}

	if lastMA10 > lastMA20 {
		signals += 1
	} else {
		signals -= 1
	}

	if lastMA20 > lastMA50 {
		signals += 1
	} else {
		signals -= 1
	}

	score := signals / 4.0
	trend := ta.classifyMAScore(score)

	return types.MAAnalysis{
		MA5:          lastMA5,
		MA10:         lastMA10,
		MA20:         lastMA20,
		MA50:         lastMA50,
		MA200:        lastMA200,
		CurrentPrice: currentPrice,
		Trend:        trend,
		Score:        score,
	}
}

// classifyMAScore classifies the MA score into a trend
func (ta *TrendAnalyzer) classifyMAScore(score float64) types.TrendDirection {
	if score > 0.75 {
		return types.StrongUptrend
	} else if score > 0.25 {
		return types.Uptrend
	} else if score > -0.25 {
		return types.Sideways
	} else if score > -0.75 {
		return types.Downtrend
	}
	return types.StrongDowntrend
}

// analyzeMomentum analyzes momentum indicators
func (ta *TrendAnalyzer) analyzeMomentum(rsi float64) types.MomentumAnalysis {
	momentum := "中性"
	if rsi > 70 {
		momentum = "超买"
	} else if rsi > 60 {
		momentum = "强势"
	} else if rsi < 30 {
		momentum = "超卖"
	} else if rsi < 40 {
		momentum = "弱势"
	}

	return types.MomentumAnalysis{
		RSI:      rsi,
		Momentum: momentum,
	}
}

// analyzeTrendStrength analyzes trend strength
func (ta *TrendAnalyzer) analyzeTrendStrength(adx float64) types.TrendStrengthAnalysis {
	strength := types.NoTrend
	if adx > 50 {
		strength = types.VeryStrong
	} else if adx > 35 {
		strength = types.Strong
	} else if adx > 20 {
		strength = types.Moderate
	} else if adx > 10 {
		strength = types.Weak
	}

	return types.TrendStrengthAnalysis{
		ADX:      adx,
		Strength: strength,
	}
}

// determineOverallTrend determines the overall trend based on all indicators
func (ta *TrendAnalyzer) determineOverallTrend(ma types.MAAnalysis, macd types.MACDAnalysis, 
	momentum types.MomentumAnalysis) (types.TrendDirection, float64) {
	
	trendScore := 0.0

	// MA contribution
	if ma.Trend == types.StrongUptrend || ma.Trend == types.Uptrend {
		trendScore += 2
	} else if ma.Trend == types.StrongDowntrend || ma.Trend == types.Downtrend {
		trendScore -= 2
	}

	// MACD contribution
	if macd.Trend == "看涨" {
		trendScore += 1
	} else if macd.Trend == "看跌" {
		trendScore -= 1
	}

	// Momentum contribution
	if momentum.Momentum == "超买" || momentum.Momentum == "强势" {
		trendScore += 1
	} else if momentum.Momentum == "超卖" || momentum.Momentum == "弱势" {
		trendScore -= 1
	}

	// Determine overall trend
	var overallTrend types.TrendDirection
	if trendScore >= 3 {
		overallTrend = types.StrongUptrend
	} else if trendScore >= 1 {
		overallTrend = types.Uptrend
	} else if trendScore > -1 {
		overallTrend = types.Sideways
	} else if trendScore > -3 {
		overallTrend = types.Downtrend
	} else {
		overallTrend = types.StrongDowntrend
	}

	return overallTrend, trendScore
}

// Helper functions
func extractCloses(data []types.OHLCV) []float64 {
	closes := make([]float64, len(data))
	for i, candle := range data {
		closes[i] = candle.Close
	}
	return closes
}

func extractHighs(data []types.OHLCV) []float64 {
	highs := make([]float64, len(data))
	for i, candle := range data {
		highs[i] = candle.High
	}
	return highs
}

func extractLows(data []types.OHLCV) []float64 {
	lows := make([]float64, len(data))
	for i, candle := range data {
		lows[i] = candle.Low
	}
	return lows
}

func extractVolumes(data []types.OHLCV) []float64 {
	volumes := make([]float64, len(data))
	for i, candle := range data {
		volumes[i] = candle.Volume
	}
	return volumes
}

func getLastValue(slice []float64, minLength int) float64 {
	if len(slice) >= minLength {
		return slice[len(slice)-1]
	}
	return 0.0
}