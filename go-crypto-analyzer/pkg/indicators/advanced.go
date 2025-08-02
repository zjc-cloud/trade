package indicators

import (
	"math"
)

// StochasticRSI 计算随机RSI
func (ti *TechnicalIndicators) StochasticRSI(data []float64, rsiPeriod, stochPeriod, k, d int) (float64, float64) {
	if len(data) < rsiPeriod+stochPeriod {
		return 50.0, 50.0
	}
	
	// 先计算RSI序列
	rsiValues := make([]float64, 0)
	for i := rsiPeriod; i <= len(data); i++ {
		rsi := ti.RSI(data[:i], rsiPeriod)
		rsiValues = append(rsiValues, rsi)
	}
	
	if len(rsiValues) < stochPeriod {
		return 50.0, 50.0
	}
	
	// 计算最近stochPeriod期的RSI最高和最低值
	startIdx := len(rsiValues) - stochPeriod
	minRSI := rsiValues[startIdx]
	maxRSI := rsiValues[startIdx]
	
	for i := startIdx; i < len(rsiValues); i++ {
		if rsiValues[i] < minRSI {
			minRSI = rsiValues[i]
		}
		if rsiValues[i] > maxRSI {
			maxRSI = rsiValues[i]
		}
	}
	
	currentRSI := rsiValues[len(rsiValues)-1]
	
	// 计算StochRSI
	var stochRSI float64
	if maxRSI-minRSI != 0 {
		stochRSI = ((currentRSI - minRSI) / (maxRSI - minRSI)) * 100
	} else {
		stochRSI = 50.0
	}
	
	// 简单返回K值，D值需要更多历史数据
	return stochRSI, stochRSI
}

// WilliamsR 威廉指标
func (ti *TechnicalIndicators) WilliamsR(high, low, close []float64, period int) float64 {
	if len(high) < period || len(low) < period || len(close) < period {
		return -50.0
	}
	
	// 找出period期内的最高和最低价
	startIdx := len(high) - period
	highest := high[startIdx]
	lowest := low[startIdx]
	
	for i := startIdx; i < len(high); i++ {
		if high[i] > highest {
			highest = high[i]
		}
		if low[i] < lowest {
			lowest = low[i]
		}
	}
	
	currentClose := close[len(close)-1]
	
	// 计算威廉指标
	if highest-lowest != 0 {
		return -100 * (highest - currentClose) / (highest - lowest)
	}
	
	return -50.0
}

// OBV 能量潮指标
func (ti *TechnicalIndicators) OBV(close, volume []float64) []float64 {
	if len(close) != len(volume) || len(close) < 2 {
		return make([]float64, len(close))
	}
	
	obv := make([]float64, len(close))
	obv[0] = volume[0]
	
	for i := 1; i < len(close); i++ {
		if close[i] > close[i-1] {
			obv[i] = obv[i-1] + volume[i]
		} else if close[i] < close[i-1] {
			obv[i] = obv[i-1] - volume[i]
		} else {
			obv[i] = obv[i-1]
		}
	}
	
	return obv
}

// ATR 平均真实波幅
func (ti *TechnicalIndicators) ATR(high, low, close []float64, period int) float64 {
	if len(high) < period+1 || len(low) < period+1 || len(close) < period+1 {
		return 0.0
	}
	
	// 计算真实波幅
	tr := make([]float64, len(high))
	for i := 1; i < len(high); i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}
	
	// 计算ATR
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	atr := sum / float64(period)
	
	// 平滑计算
	for i := period + 1; i < len(tr); i++ {
		atr = (atr*float64(period-1) + tr[i]) / float64(period)
	}
	
	return atr
}

// CCI 商品通道指数
func (ti *TechnicalIndicators) CCI(high, low, close []float64, period int) float64 {
	if len(high) < period || len(low) < period || len(close) < period {
		return 0.0
	}
	
	// 计算典型价格
	tp := make([]float64, len(high))
	for i := 0; i < len(high); i++ {
		tp[i] = (high[i] + low[i] + close[i]) / 3
	}
	
	// 计算移动平均
	ma := ti.SMA(tp, period)
	if len(ma) == 0 {
		return 0.0
	}
	currentMA := ma[len(ma)-1]
	
	// 计算平均偏差
	sum := 0.0
	startIdx := len(tp) - period
	for i := startIdx; i < len(tp); i++ {
		sum += math.Abs(tp[i] - currentMA)
	}
	meanDev := sum / float64(period)
	
	// 计算CCI
	currentTP := tp[len(tp)-1]
	if meanDev != 0 {
		return (currentTP - currentMA) / (0.015 * meanDev)
	}
	
	return 0.0
}