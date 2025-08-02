package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"
	
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// Exporter 数据导出器
type Exporter struct {
	format string
}

// NewExporter 创建导出器
func NewExporter(format string) *Exporter {
	return &Exporter{format: format}
}

// ExportAnalysis 导出分析结果
func (e *Exporter) ExportAnalysis(analysis *types.Analysis, evidences []types.Evidence) error {
	filename := fmt.Sprintf("analysis_%s_%s.%s", 
		analysis.Symbol, 
		time.Now().Format("20060102_150405"), 
		e.format)
	
	switch e.format {
	case "json":
		return e.exportJSON(filename, analysis, evidences)
	case "csv":
		return e.exportCSV(filename, analysis, evidences)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

// exportJSON 导出为JSON格式
func (e *Exporter) exportJSON(filename string, analysis *types.Analysis, evidences []types.Evidence) error {
	data := map[string]interface{}{
		"timestamp": time.Now(),
		"analysis":  analysis,
		"evidences": evidences,
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportCSV 导出为CSV格式
func (e *Exporter) exportCSV(filename string, analysis *types.Analysis, evidences []types.Evidence) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// 写入头部
	headers := []string{
		"时间", "交易对", "价格", "趋势", "趋势得分",
		"RSI", "MACD", "ADX", "成交量比",
		"MA5", "MA20", "MA50",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}
	
	// 写入数据
	row := []string{
		analysis.Timestamp.Format("2006-01-02 15:04:05"),
		analysis.Symbol,
		fmt.Sprintf("%.2f", analysis.CurrentPrice),
		string(analysis.OverallTrend),
		fmt.Sprintf("%.2f", analysis.TrendScore),
		fmt.Sprintf("%.1f", analysis.Momentum.RSI),
		fmt.Sprintf("%.2f", analysis.MACDAnalysis.MACD),
		fmt.Sprintf("%.1f", analysis.TrendStrength.ADX),
		fmt.Sprintf("%.2f", analysis.Volume.VolumeRatio),
		fmt.Sprintf("%.2f", analysis.MAAnalysis.MA5),
		fmt.Sprintf("%.2f", analysis.MAAnalysis.MA20),
		fmt.Sprintf("%.2f", analysis.MAAnalysis.MA50),
	}
	
	return writer.Write(row)
}

// ExportOHLCV 导出K线数据
func (e *Exporter) ExportOHLCV(symbol string, data []types.OHLCV) error {
	filename := fmt.Sprintf("ohlcv_%s_%s.csv", 
		symbol, 
		time.Now().Format("20060102_150405"))
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// 写入头部
	headers := []string{"时间", "开盘", "最高", "最低", "收盘", "成交量"}
	if err := writer.Write(headers); err != nil {
		return err
	}
	
	// 写入数据
	for _, candle := range data {
		row := []string{
			candle.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", candle.Open),
			fmt.Sprintf("%.2f", candle.High),
			fmt.Sprintf("%.2f", candle.Low),
			fmt.Sprintf("%.2f", candle.Close),
			fmt.Sprintf("%.0f", candle.Volume),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	
	return nil
}