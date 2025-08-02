package analysis

import (
	"math"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// DynamicAnalyzer 提供动态权重的分析系统
type DynamicAnalyzer struct {
	// 权重配置
	weights map[string]float64
	// 市场状态
	marketCondition string
}

// NewDynamicAnalyzer 创建动态分析器
func NewDynamicAnalyzer() *DynamicAnalyzer {
	return &DynamicAnalyzer{
		weights: make(map[string]float64),
		marketCondition: "normal",
	}
}

// 根据市场状态动态调整权重
func (da *DynamicAnalyzer) AdjustWeights(volatility float64, volume types.VolumeAnalysis, adx float64) {
	// 高波动市场
	if volatility > 0.05 { // 5%波动率
		da.marketCondition = "high_volatility"
		da.weights["MA"] = 0.15      // 降低MA权重
		da.weights["MACD"] = 0.25    // 提高动量权重
		da.weights["RSI"] = 0.30     // 提高超买超卖权重
		da.weights["Volume"] = 0.30  // 提高成交量权重
		
	// 趋势市场
	} else if adx > 35 {
		da.marketCondition = "trending"
		da.weights["MA"] = 0.35      // 提高MA权重
		da.weights["MACD"] = 0.30    // 动量重要
		da.weights["RSI"] = 0.15     // 降低RSI权重
		da.weights["Volume"] = 0.20
		
	// 震荡市场
	} else {
		da.marketCondition = "ranging"
		da.weights["MA"] = 0.20
		da.weights["MACD"] = 0.20
		da.weights["RSI"] = 0.35     // 提高RSI权重
		da.weights["Volume"] = 0.25
	}
}

// 计算指标可信度
func (da *DynamicAnalyzer) CalculateConfidence(indicators map[string]bool) float64 {
	// 统计一致性
	agreeCount := 0
	totalCount := 0
	
	for _, bullish := range indicators {
		totalCount++
		if bullish {
			agreeCount++
		}
	}
	
	// 一致性越高，可信度越高
	agreement := float64(agreeCount) / float64(totalCount)
	if agreement > 0.5 {
		agreement = 1 - agreement
	}
	
	// 0.5是完全不一致，0是完全一致
	confidence := 1 - (agreement * 2)
	
	return confidence
}

// 智能证据评估
func (da *DynamicAnalyzer) EvaluateEvidence(evidence types.Evidence, context map[string]interface{}) float64 {
	baseStrength := evidence.Strength
	
	// 根据市场状态调整证据强度
	switch da.marketCondition {
	case "high_volatility":
		// 高波动时，短期指标更重要
		if evidence.Category == "移动平均线" && evidence.Description[:8] == "MA5" {
			baseStrength *= 1.5
		}
	case "trending":
		// 趋势市场，中长期指标更重要
		if evidence.Category == "移动平均线" && (evidence.Description[:9] == "MA50" || evidence.Description[:9] == "MA20") {
			baseStrength *= 1.3
		}
	case "ranging":
		// 震荡市场，超买超卖指标更重要
		if evidence.Category == "RSI" {
			baseStrength *= 1.4
		}
	}
	
	// 成交量验证
	if volumeRatio, ok := context["volumeRatio"].(float64); ok {
		if volumeRatio > 1.5 {
			// 放量验证，增强信号
			baseStrength *= 1.2
		} else if volumeRatio < 0.5 {
			// 缩量，减弱信号
			baseStrength *= 0.8
		}
	}
	
	return baseStrength
}

// 多指标融合决策
func (da *DynamicAnalyzer) FusionDecision(evidences []types.Evidence) (string, float64) {
	// 贝叶斯推理
	bullishProbability := 0.5 // 先验概率
	
	for _, evidence := range evidences {
		// 计算似然比
		likelihoodRatio := 1.0
		
		switch evidence.Type {
		case types.BullishEvidence:
			likelihoodRatio = 1 + evidence.Strength
		case types.BearishEvidence:
			likelihoodRatio = 1 / (1 + math.Abs(evidence.Strength))
		}
		
		// 更新后验概率
		bullishProbability = (bullishProbability * likelihoodRatio) / 
			(bullishProbability * likelihoodRatio + (1 - bullishProbability))
	}
	
	// 决策
	if bullishProbability > 0.7 {
		return "强烈看涨", bullishProbability
	} else if bullishProbability > 0.55 {
		return "偏多", bullishProbability
	} else if bullishProbability < 0.3 {
		return "强烈看跌", bullishProbability
	} else if bullishProbability < 0.45 {
		return "偏空", bullishProbability
	}
	
	return "中性", bullishProbability
}

// 指标冲突检测
func (da *DynamicAnalyzer) DetectConflicts(evidences []types.Evidence) []string {
	conflicts := []string{}
	
	// 检查MA和MACD是否冲突
	maSignal := ""
	macdSignal := ""
	
	for _, ev := range evidences {
		if ev.Category == "移动平均线" && ev.Description[:3] == "MA5" {
			if ev.Type == types.BullishEvidence {
				maSignal = "bullish"
			} else {
				maSignal = "bearish"
			}
		}
		if ev.Category == "MACD" {
			if ev.Type == types.BullishEvidence {
				macdSignal = "bullish"
			} else {
				macdSignal = "bearish"
			}
		}
	}
	
	if maSignal != "" && macdSignal != "" && maSignal != macdSignal {
		conflicts = append(conflicts, "MA和MACD信号冲突，谨慎操作")
	}
	
	// 检查价格和成交量是否背离
	for _, ev := range evidences {
		if ev.Category == "成交量" && len(ev.Description) >= 4 && ev.Description[:4] == "放量" {
			if ev.Type == types.BearishEvidence {
				conflicts = append(conflicts, "放量下跌，卖压沉重")
			}
		}
	}
	
	return conflicts
}