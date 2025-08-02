package analysis

import (
	"fmt"

	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// EvidenceCollector collects and analyzes trading evidence
type EvidenceCollector struct {
	evidences []types.Evidence
}

// NewEvidenceCollector creates a new EvidenceCollector
func NewEvidenceCollector() *EvidenceCollector {
	return &EvidenceCollector{
		evidences: make([]types.Evidence, 0),
	}
}

// Clear clears all collected evidence
func (ec *EvidenceCollector) Clear() {
	ec.evidences = make([]types.Evidence, 0)
}

// AddEvidence adds a new piece of evidence
func (ec *EvidenceCollector) AddEvidence(evidence types.Evidence) {
	ec.evidences = append(ec.evidences, evidence)
}

// AnalyzeMAEvidence analyzes moving average evidence
func (ec *EvidenceCollector) AnalyzeMAEvidence(ma types.MAAnalysis, currentPrice float64) {
	// Price vs MA5
	if currentPrice > ma.MA5 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "移动平均线",
			Description: fmt.Sprintf("价格(%.2f)高于MA5(%.2f)，短期趋势向上", currentPrice, ma.MA5),
			Strength:    0.3,
			Data: map[string]interface{}{
				"price": currentPrice,
				"ma5":   ma.MA5,
			},
		})
	} else {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "移动平均线",
			Description: fmt.Sprintf("价格(%.2f)低于MA5(%.2f)，短期趋势向下", currentPrice, ma.MA5),
			Strength:    -0.3,
			Data: map[string]interface{}{
				"price": currentPrice,
				"ma5":   ma.MA5,
			},
		})
	}

	// MA5 vs MA20
	if ma.MA5 > ma.MA20 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "移动平均线",
			Description: fmt.Sprintf("MA5(%.2f)高于MA20(%.2f)，中期趋势向上", ma.MA5, ma.MA20),
			Strength:    0.4,
			Data: map[string]interface{}{
				"ma5":  ma.MA5,
				"ma20": ma.MA20,
			},
		})
	} else {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "移动平均线",
			Description: fmt.Sprintf("MA5(%.2f)低于MA20(%.2f)，中期趋势向下", ma.MA5, ma.MA20),
			Strength:    -0.4,
			Data: map[string]interface{}{
				"ma5":  ma.MA5,
				"ma20": ma.MA20,
			},
		})
	}

	// Perfect alignment check
	if currentPrice > ma.MA5 && ma.MA5 > ma.MA10 && ma.MA10 > ma.MA20 && ma.MA20 > ma.MA50 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "移动平均线",
			Description: "完美多头排列：价格>MA5>MA10>MA20>MA50",
			Strength:    0.8,
			Data:        map[string]interface{}{},
		})
	} else if currentPrice < ma.MA5 && ma.MA5 < ma.MA10 && ma.MA10 < ma.MA20 && ma.MA20 < ma.MA50 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "移动平均线",
			Description: "完美空头排列：价格<MA5<MA10<MA20<MA50",
			Strength:    -0.8,
			Data:        map[string]interface{}{},
		})
	}
}

// AnalyzeMACDEvidence analyzes MACD evidence
func (ec *EvidenceCollector) AnalyzeMACDEvidence(macd types.MACDAnalysis) {
	// MACD vs Signal
	if macd.MACD > macd.Signal {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "MACD",
			Description: fmt.Sprintf("MACD(%.2f)高于Signal(%.2f)，动量向上", macd.MACD, macd.Signal),
			Strength:    0.5,
			Data: map[string]interface{}{
				"macd":   macd.MACD,
				"signal": macd.Signal,
			},
		})
	} else {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "MACD",
			Description: fmt.Sprintf("MACD(%.2f)低于Signal(%.2f)，动量向下", macd.MACD, macd.Signal),
			Strength:    -0.5,
			Data: map[string]interface{}{
				"macd":   macd.MACD,
				"signal": macd.Signal,
			},
		})
	}

	// Histogram
	if macd.Histogram > 0 && macd.Histogram > macd.MACD*0.1 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "MACD",
			Description: fmt.Sprintf("MACD柱状图为正(%.2f)且较大，买入动量强", macd.Histogram),
			Strength:    0.4,
			Data:        map[string]interface{}{"histogram": macd.Histogram},
		})
	} else if macd.Histogram < 0 && -macd.Histogram > -macd.MACD*0.1 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "MACD",
			Description: fmt.Sprintf("MACD柱状图为负(%.2f)且较大，卖出动量强", macd.Histogram),
			Strength:    -0.4,
			Data:        map[string]interface{}{"histogram": macd.Histogram},
		})
	}
}

// AnalyzeRSIEvidence analyzes RSI evidence
func (ec *EvidenceCollector) AnalyzeRSIEvidence(rsi float64) {
	if rsi > 70 {
		ec.AddEvidence(types.Evidence{
			Type:        types.WarningEvidence,
			Category:    "RSI",
			Description: fmt.Sprintf("RSI(%.2f)>70，处于超买区域，可能回调", rsi),
			Strength:    -0.3,
			Data:        map[string]interface{}{"rsi": rsi},
		})
	} else if rsi > 60 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "RSI",
			Description: fmt.Sprintf("RSI(%.2f)处于强势区域(60-70)，上涨动能充足", rsi),
			Strength:    0.3,
			Data:        map[string]interface{}{"rsi": rsi},
		})
	} else if rsi < 30 {
		ec.AddEvidence(types.Evidence{
			Type:        types.WarningEvidence,
			Category:    "RSI",
			Description: fmt.Sprintf("RSI(%.2f)<30，处于超卖区域，可能反弹", rsi),
			Strength:    0.3,
			Data:        map[string]interface{}{"rsi": rsi},
		})
	} else if rsi < 40 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "RSI",
			Description: fmt.Sprintf("RSI(%.2f)处于弱势区域(30-40)，下跌动能较强", rsi),
			Strength:    -0.3,
			Data:        map[string]interface{}{"rsi": rsi},
		})
	} else {
		ec.AddEvidence(types.Evidence{
			Type:        types.NeutralEvidence,
			Category:    "RSI",
			Description: fmt.Sprintf("RSI(%.2f)处于中性区域(40-60)", rsi),
			Strength:    0,
			Data:        map[string]interface{}{"rsi": rsi},
		})
	}
}

// AnalyzeVolumeEvidence analyzes volume evidence
func (ec *EvidenceCollector) AnalyzeVolumeEvidence(volume types.VolumeAnalysis, priceChange float64) {
	if volume.VolumeRatio > 2 && priceChange > 0 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "成交量",
			Description: fmt.Sprintf("放量上涨：成交量是均量的%.1f倍，买入意愿强烈", volume.VolumeRatio),
			Strength:    0.6,
			Data:        map[string]interface{}{"volumeRatio": volume.VolumeRatio},
		})
	} else if volume.VolumeRatio > 2 && priceChange < 0 {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "成交量",
			Description: fmt.Sprintf("放量下跌：成交量是均量的%.1f倍，卖出压力大", volume.VolumeRatio),
			Strength:    -0.6,
			Data:        map[string]interface{}{"volumeRatio": volume.VolumeRatio},
		})
	} else if volume.VolumeRatio < 0.5 && priceChange > 0 {
		ec.AddEvidence(types.Evidence{
			Type:        types.WarningEvidence,
			Category:    "成交量",
			Description: fmt.Sprintf("缩量上涨：成交量仅为均量的%.1f倍，上涨缺乏支撑", volume.VolumeRatio),
			Strength:    -0.2,
			Data:        map[string]interface{}{"volumeRatio": volume.VolumeRatio},
		})
	} else if volume.VolumeRatio < 0.5 && priceChange < 0 {
		ec.AddEvidence(types.Evidence{
			Type:        types.NeutralEvidence,
			Category:    "成交量",
			Description: fmt.Sprintf("缩量下跌：成交量仅为均量的%.1f倍，抛压减轻", volume.VolumeRatio),
			Strength:    0.2,
			Data:        map[string]interface{}{"volumeRatio": volume.VolumeRatio},
		})
	}
}

// AnalyzeSREvidence analyzes support and resistance evidence
func (ec *EvidenceCollector) AnalyzeSREvidence(currentPrice float64, sr types.SRAnalysis) {
	r1 := sr.Resistance["R1"]
	s1 := sr.Support["S1"]
	pivot := sr.Pivot

	// Distance to levels
	distanceToR1 := (r1 - currentPrice) / currentPrice * 100
	distanceToS1 := (currentPrice - s1) / currentPrice * 100

	if distanceToR1 < 1 {
		ec.AddEvidence(types.Evidence{
			Type:        types.WarningEvidence,
			Category:    "支撑阻力",
			Description: fmt.Sprintf("接近阻力位R1(%.2f)，上涨空间有限(%.1f%%)", r1, distanceToR1),
			Strength:    -0.3,
			Data:        map[string]interface{}{"r1": r1, "distance": distanceToR1},
		})
	}

	if distanceToS1 < 1 {
		ec.AddEvidence(types.Evidence{
			Type:        types.WarningEvidence,
			Category:    "支撑阻力",
			Description: fmt.Sprintf("接近支撑位S1(%.2f)，下跌空间有限(%.1f%%)", s1, distanceToS1),
			Strength:    0.3,
			Data:        map[string]interface{}{"s1": s1, "distance": distanceToS1},
		})
	}

	if currentPrice > pivot {
		ec.AddEvidence(types.Evidence{
			Type:        types.BullishEvidence,
			Category:    "支撑阻力",
			Description: fmt.Sprintf("价格(%.2f)高于轴心点(%.2f)，多头占优", currentPrice, pivot),
			Strength:    0.2,
			Data:        map[string]interface{}{"price": currentPrice, "pivot": pivot},
		})
	} else {
		ec.AddEvidence(types.Evidence{
			Type:        types.BearishEvidence,
			Category:    "支撑阻力",
			Description: fmt.Sprintf("价格(%.2f)低于轴心点(%.2f)，空头占优", currentPrice, pivot),
			Strength:    -0.2,
			Data:        map[string]interface{}{"price": currentPrice, "pivot": pivot},
		})
	}
}

// GetSummary returns a summary of all collected evidence
func (ec *EvidenceCollector) GetSummary() map[string]interface{} {
	bullishCount := 0
	bearishCount := 0
	warningCount := 0
	totalStrength := 0.0

	var bullishEvidences []types.Evidence
	var bearishEvidences []types.Evidence
	var warningEvidences []types.Evidence

	for _, evidence := range ec.evidences {
		switch evidence.Type {
		case types.BullishEvidence:
			bullishCount++
			bullishEvidences = append(bullishEvidences, evidence)
		case types.BearishEvidence:
			bearishCount++
			bearishEvidences = append(bearishEvidences, evidence)
		case types.WarningEvidence:
			warningCount++
			warningEvidences = append(warningEvidences, evidence)
		}
		totalStrength += evidence.Strength
	}

	return map[string]interface{}{
		"bullishCount":     bullishCount,
		"bearishCount":     bearishCount,
		"warningCount":     warningCount,
		"totalStrength":    totalStrength,
		"bullishEvidences": bullishEvidences,
		"bearishEvidences": bearishEvidences,
		"warningEvidences": warningEvidences,
		"allEvidences":     ec.evidences,
	}
}