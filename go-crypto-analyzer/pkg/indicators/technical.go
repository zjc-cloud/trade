package indicators

import (
	"math"

	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// TechnicalIndicators provides methods for calculating technical indicators
type TechnicalIndicators struct{}

// NewTechnicalIndicators creates a new TechnicalIndicators instance
func NewTechnicalIndicators() *TechnicalIndicators {
	return &TechnicalIndicators{}
}

// SMA calculates Simple Moving Average
func (ti *TechnicalIndicators) SMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}

	sma := make([]float64, len(data))
	sum := 0.0

	// Calculate first SMA
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	sma[period-1] = sum / float64(period)

	// Calculate remaining SMAs
	for i := period; i < len(data); i++ {
		sum = sum - data[i-period] + data[i]
		sma[i] = sum / float64(period)
	}

	return sma
}

// EMA calculates Exponential Moving Average
func (ti *TechnicalIndicators) EMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}

	ema := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	// Calculate initial SMA for first EMA value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA
	for i := period; i < len(data); i++ {
		ema[i] = (data[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// MACD calculates MACD indicator
func (ti *TechnicalIndicators) MACD(data []float64, fast, slow, signal int) types.MACDAnalysis {
	if len(data) < slow {
		return types.MACDAnalysis{}
	}

	emaFast := ti.EMA(data, fast)
	emaSlow := ti.EMA(data, slow)

	// Calculate MACD line
	macdLine := make([]float64, len(data))
	for i := slow - 1; i < len(data); i++ {
		macdLine[i] = emaFast[i] - emaSlow[i]
	}

	// Calculate signal line
	signalLine := ti.EMA(macdLine[slow-1:], signal)

	// Get latest values
	latestMACD := macdLine[len(macdLine)-1]
	latestSignal := 0.0
	if len(signalLine) > 0 {
		latestSignal = signalLine[len(signalLine)-1]
	}
	histogram := latestMACD - latestSignal

	// Determine trend
	trend := "中性"
	if latestMACD > latestSignal {
		if histogram > 0 {
			trend = "看涨"
		}
	} else {
		if histogram < 0 {
			trend = "看跌"
		}
	}

	return types.MACDAnalysis{
		MACD:       latestMACD,
		Signal:     latestSignal,
		Histogram:  histogram,
		Trend:      trend,
		Divergence: "无背离",
	}
}

// RSI calculates Relative Strength Index
func (ti *TechnicalIndicators) RSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 50.0
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial average gain and loss
	for i := 1; i <= period; i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate subsequent values using smoothed average
	for i := period + 1; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			avgGain = (avgGain*(float64(period)-1) + change) / float64(period)
			avgLoss = (avgLoss * (float64(period) - 1)) / float64(period)
		} else {
			avgGain = (avgGain * (float64(period) - 1)) / float64(period)
			avgLoss = (avgLoss*(float64(period)-1) + math.Abs(change)) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi
}

// BollingerBands calculates Bollinger Bands
func (ti *TechnicalIndicators) BollingerBands(data []float64, period int, stdDev float64) (upper, middle, lower []float64) {
	middle = ti.SMA(data, period)
	upper = make([]float64, len(data))
	lower = make([]float64, len(data))

	for i := period - 1; i < len(data); i++ {
		// Calculate standard deviation
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := data[j] - middle[i]
			sum += diff * diff
		}
		std := math.Sqrt(sum / float64(period))

		upper[i] = middle[i] + (std * stdDev)
		lower[i] = middle[i] - (std * stdDev)
	}

	return upper, middle, lower
}

// ADX calculates Average Directional Index
func (ti *TechnicalIndicators) ADX(high, low, close []float64, period int) float64 {
	if len(high) < period*2 || len(low) < period*2 || len(close) < period*2 {
		return 0.0
	}

	// Calculate True Range
	tr := make([]float64, len(high))
	for i := 1; i < len(high); i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Calculate directional movements
	plusDM := make([]float64, len(high))
	minusDM := make([]float64, len(high))

	for i := 1; i < len(high); i++ {
		upMove := high[i] - high[i-1]
		downMove := low[i-1] - low[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	// Calculate smoothed values
	atr := ti.SMA(tr[1:], period)
	plusDI := make([]float64, len(plusDM))
	minusDI := make([]float64, len(minusDM))

	if len(atr) > 0 && atr[len(atr)-1] != 0 {
		smoothedPlusDM := ti.SMA(plusDM[1:], period)
		smoothedMinusDM := ti.SMA(minusDM[1:], period)

		for i := 0; i < len(smoothedPlusDM); i++ {
			if i < len(atr) && atr[i] != 0 {
				plusDI[i] = 100 * smoothedPlusDM[i] / atr[i]
				minusDI[i] = 100 * smoothedMinusDM[i] / atr[i]
			}
		}
	}

	// Calculate DX and ADX
	dx := make([]float64, len(plusDI))
	for i := 0; i < len(plusDI); i++ {
		sum := plusDI[i] + minusDI[i]
		if sum != 0 {
			dx[i] = 100 * math.Abs(plusDI[i]-minusDI[i]) / sum
		}
	}

	adxValues := ti.SMA(dx, period)
	if len(adxValues) > 0 {
		return adxValues[len(adxValues)-1]
	}

	return 0.0
}

// VolumeAnalysis analyzes volume patterns
func (ti *TechnicalIndicators) VolumeAnalysis(volume []float64, period int) types.VolumeAnalysis {
	if len(volume) < period {
		return types.VolumeAnalysis{
			CurrentVolume: 0,
			VolumeMA:      0,
			VolumeRatio:   0,
			VolumeTrend:   "数据不足",
		}
	}

	volumeMA := ti.SMA(volume, period)
	currentVolume := volume[len(volume)-1]
	avgVolume := volumeMA[len(volumeMA)-1]

	volumeRatio := 0.0
	if avgVolume > 0 {
		volumeRatio = currentVolume / avgVolume
	}

	volumeTrend := "正常量能"
	if volumeRatio > 2 {
		volumeTrend = "放量"
	} else if volumeRatio < 0.5 {
		volumeTrend = "缩量"
	}

	return types.VolumeAnalysis{
		CurrentVolume: currentVolume,
		VolumeMA:      avgVolume,
		VolumeRatio:   volumeRatio,
		VolumeTrend:   volumeTrend,
	}
}

// PivotPoints calculates pivot points for support and resistance
func (ti *TechnicalIndicators) PivotPoints(high, low, close float64) types.SRAnalysis {
	pivot := (high + low + close) / 3

	r1 := 2*pivot - low
	r2 := pivot + (high - low)
	r3 := high + 2*(pivot-low)

	s1 := 2*pivot - high
	s2 := pivot - (high - low)
	s3 := low - 2*(high-pivot)

	return types.SRAnalysis{
		Pivot: pivot,
		Resistance: map[string]float64{
			"R1": r1,
			"R2": r2,
			"R3": r3,
		},
		Support: map[string]float64{
			"S1": s1,
			"S2": s2,
			"S3": s3,
		},
	}
}