package indicators

import (
	"math"
	"testing"
)

func TestSMA(t *testing.T) {
	ti := NewTechnicalIndicators()
	
	// 测试数据
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	// 计算5期SMA
	sma := ti.SMA(data, 5)
	
	// 验证最后一个值
	expected := (6.0 + 7.0 + 8.0 + 9.0 + 10.0) / 5.0
	if math.Abs(sma[len(sma)-1]-expected) > 0.001 {
		t.Errorf("SMA calculation error: expected %.2f, got %.2f", expected, sma[len(sma)-1])
	}
}

func TestRSI(t *testing.T) {
	ti := NewTechnicalIndicators()
	
	// 测试数据 - 持续上涨
	upData := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115}
	rsi := ti.RSI(upData, 14)
	
	// RSI应该接近100（强烈超买）
	if rsi < 90 {
		t.Errorf("RSI for uptrend should be > 90, got %.2f", rsi)
	}
	
	// 测试数据 - 持续下跌
	downData := []float64{115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100}
	rsi = ti.RSI(downData, 14)
	
	// RSI应该接近0（强烈超卖）
	if rsi > 10 {
		t.Errorf("RSI for downtrend should be < 10, got %.2f", rsi)
	}
}

func TestMACD(t *testing.T) {
	ti := NewTechnicalIndicators()
	
	// 生成测试数据
	data := make([]float64, 50)
	for i := 0; i < 50; i++ {
		data[i] = 100 + float64(i)*0.5 // 稳定上升趋势
	}
	
	macd := ti.MACD(data, 12, 26, 9)
	
	// 在上升趋势中，MACD应该为正
	if macd.MACD < 0 {
		t.Errorf("MACD should be positive in uptrend, got %.2f", macd.MACD)
	}
	
	// 趋势应该是看涨
	if macd.Trend != "看涨" {
		t.Errorf("MACD trend should be bullish, got %s", macd.Trend)
	}
}

func TestPivotPoints(t *testing.T) {
	ti := NewTechnicalIndicators()
	
	high := 110.0
	low := 100.0
	close := 105.0
	
	sr := ti.PivotPoints(high, low, close)
	
	// 验证轴心点
	expectedPivot := (high + low + close) / 3
	if math.Abs(sr.Pivot-expectedPivot) > 0.001 {
		t.Errorf("Pivot calculation error: expected %.2f, got %.2f", expectedPivot, sr.Pivot)
	}
	
	// 验证阻力位
	if sr.Resistance["R1"] <= sr.Pivot {
		t.Error("R1 should be above pivot")
	}
	
	// 验证支撑位
	if sr.Support["S1"] >= sr.Pivot {
		t.Error("S1 should be below pivot")
	}
}