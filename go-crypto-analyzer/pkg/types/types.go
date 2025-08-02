package types

import (
	"time"
)

// OHLCV represents a candlestick data point
type OHLCV struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	StrongUptrend   TrendDirection = "强劲上涨趋势"
	Uptrend         TrendDirection = "上涨趋势"
	Sideways        TrendDirection = "横盘震荡"
	Downtrend       TrendDirection = "下跌趋势"
	StrongDowntrend TrendDirection = "强劲下跌趋势"
)

// TrendStrength represents the strength of a trend
type TrendStrength string

const (
	VeryStrong TrendStrength = "非常强"
	Strong     TrendStrength = "强"
	Moderate   TrendStrength = "中等"
	Weak       TrendStrength = "弱"
	NoTrend    TrendStrength = "无趋势"
)

// Analysis represents the complete analysis result
type Analysis struct {
	Symbol          string
	CurrentPrice    float64
	Timestamp       time.Time
	OverallTrend    TrendDirection
	TrendScore      float64
	MAAnalysis      MAAnalysis
	MACDAnalysis    MACDAnalysis
	Momentum        MomentumAnalysis
	TrendStrength   TrendStrengthAnalysis
	Volume          VolumeAnalysis
	SupportResistance SRAnalysis
}

// MAAnalysis represents moving average analysis
type MAAnalysis struct {
	MA5          float64
	MA10         float64
	MA20         float64
	MA50         float64
	MA200        float64
	CurrentPrice float64
	Trend        TrendDirection
	Score        float64
}

// MACDAnalysis represents MACD analysis
type MACDAnalysis struct {
	MACD       float64
	Signal     float64
	Histogram  float64
	Trend      string
	Divergence string
}

// MomentumAnalysis represents momentum indicators
type MomentumAnalysis struct {
	RSI      float64
	Momentum string
}

// TrendStrengthAnalysis represents trend strength
type TrendStrengthAnalysis struct {
	ADX      float64
	Strength TrendStrength
}

// VolumeAnalysis represents volume analysis
type VolumeAnalysis struct {
	CurrentVolume float64
	VolumeMA      float64
	VolumeRatio   float64
	VolumeTrend   string
}

// SRAnalysis represents support and resistance analysis
type SRAnalysis struct {
	Pivot      float64
	Resistance map[string]float64
	Support    map[string]float64
}

// Evidence represents a piece of analysis evidence
type Evidence struct {
	Type        EvidenceType
	Category    string
	Description string
	Strength    float64
	Data        map[string]interface{}
}

// EvidenceType represents the type of evidence
type EvidenceType string

const (
	BullishEvidence EvidenceType = "看涨证据"
	BearishEvidence EvidenceType = "看跌证据"
	NeutralEvidence EvidenceType = "中性证据"
	WarningEvidence EvidenceType = "警告信号"
)

// CryptoConfig represents cryptocurrency configuration
type CryptoConfig struct {
	Symbol       string
	Name         string
	Category     string
	Weight       float64
	Timeframes   []string
	KeyLevels    KeyLevels
}

// KeyLevels represents important price levels
type KeyLevels struct {
	Psychological        []float64
	HistoricalSupport    []float64
	HistoricalResistance []float64
}

// FearGreedIndex represents the fear and greed index
type FearGreedIndex struct {
	Value          int
	Classification string
	Sentiment      string
	Timestamp      time.Time
}