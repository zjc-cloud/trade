package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/zjc/go-crypto-analyzer/internal/config"
	"github.com/zjc/go-crypto-analyzer/pkg/analysis"
	"github.com/zjc/go-crypto-analyzer/pkg/data"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

var (
	symbols    []string
	watchlist  string
	interval   string
	limit      int
	useYahoo   bool
	continuous bool
	delay      int
)

var rootCmd = &cobra.Command{
	Use:   "crypto-analyzer",
	Short: "加密货币市场趋势分析工具",
	Long:  `使用Go语言开发的加密货币市场技术分析工具，支持多种技术指标和实时监控。`,
	Run:   runAnalysis,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&symbols, "symbols", "s", []string{}, "要分析的交易对列表")
	rootCmd.Flags().StringVarP(&watchlist, "watchlist", "w", "top3", "使用预设的监控列表")
	rootCmd.Flags().StringVarP(&interval, "interval", "i", "1h", "K线时间间隔 (15m/30m/1h/4h/1d)")
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 100, "获取K线数量")
	rootCmd.Flags().BoolVarP(&useYahoo, "yahoo", "y", false, "使用Yahoo Finance数据源")
	rootCmd.Flags().BoolVarP(&continuous, "continuous", "c", false, "持续监控模式")
	rootCmd.Flags().IntVarP(&delay, "delay", "d", 300, "监控间隔（秒）")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
	// Determine symbols to analyze
	symbolsToAnalyze := symbols
	if len(symbolsToAnalyze) == 0 {
		symbolsToAnalyze = config.GetWatchlist(watchlist)
	}

	// Create data fetcher
	var fetcher data.Fetcher
	if useYahoo {
		fetcher = data.NewYahooFinanceFetcher()
		fmt.Println("使用Yahoo Finance数据源")
	} else {
		fetcher = data.NewBinanceFetcher()
		fmt.Println("使用Binance数据源")
	}

	// Create analyzers
	trendAnalyzer := analysis.NewTrendAnalyzer()
	evidenceCollector := analysis.NewEvidenceCollector()

	// Fetch Fear & Greed Index
	fgFetcher := data.NewFearGreedFetcher()
	fearGreed, err := fgFetcher.Fetch()
	if err == nil {
		printFearGreedIndex(fearGreed)
	}

	// Analysis loop
	for {
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Printf("🚀 加密货币市场分析 - %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("📊 时间周期: %s | 数据点: %d\n", interval, limit)
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		for _, symbol := range symbolsToAnalyze {
			analyzeSymbol(symbol, fetcher, trendAnalyzer, evidenceCollector)
		}

		if !continuous {
			break
		}

		fmt.Printf("\n⏰ 下次更新: %d秒后\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func analyzeSymbol(symbol string, fetcher data.Fetcher, analyzer *analysis.TrendAnalyzer, collector *analysis.EvidenceCollector) {
	fmt.Printf("\n📊 分析 %s\n", color.YellowString(symbol))
	fmt.Println(strings.Repeat("-", 60))

	// Fetch OHLCV data
	ohlcv, err := fetcher.FetchOHLCV(symbol, interval, limit)
	if err != nil {
		// 提供更友好的错误信息
		if strings.Contains(err.Error(), "418") || strings.Contains(err.Error(), "banned") {
			color.Red("  ❌ API访问被限制，请稍后再试或使用 -y 参数切换到Yahoo数据源")
		} else if strings.Contains(err.Error(), "network") || strings.Contains(err.Error(), "connection") {
			color.Red("  ❌ 网络连接失败，请检查网络连接")
		} else {
			color.Red("  ❌ 获取数据失败: %v", err)
		}
		fmt.Println("  💡 提示: 可以尝试以下操作:")
		fmt.Println("     1. 使用 -y 参数切换到Yahoo Finance数据源")
		fmt.Println("     2. 减少请求频率或数据量")
		fmt.Println("     3. 检查交易对名称是否正确")
		return
	}

	if len(ohlcv) < 50 {
		color.Red("  ❌ 数据不足（需要至少50根K线）")
		return
	}

	// Perform analysis
	result, err := analyzer.AnalyzeComprehensive(ohlcv)
	if err != nil {
		color.Red("  ❌ 分析失败: %v", err)
		return
	}

	// Collect evidence
	collector.Clear()
	collector.AnalyzeMAEvidence(result.MAAnalysis, result.CurrentPrice)
	collector.AnalyzeMACDEvidence(result.MACDAnalysis)
	collector.AnalyzeRSIEvidence(result.Momentum.RSI)
	collector.AnalyzeSREvidence(result.CurrentPrice, result.SupportResistance)

	// Calculate price change
	priceChange := 0.0
	if len(ohlcv) > 1 {
		priceChange = (ohlcv[len(ohlcv)-1].Close - ohlcv[len(ohlcv)-2].Close) / ohlcv[len(ohlcv)-2].Close
	}
	collector.AnalyzeVolumeEvidence(result.Volume, priceChange)

	// Get evidence summary
	evidenceSummary := collector.GetSummary()

	// Print results
	printAnalysisResult(result, evidenceSummary)

	// Print price chart
	printPriceChart(ohlcv)
}

func printFearGreedIndex(fg *types.FearGreedIndex) {
	fmt.Printf("\n😱 恐慌贪婪指数: ")
	
	value := fg.Value
	var colorFunc func(format string, a ...interface{}) string
	
	if value < 25 {
		colorFunc = color.RedString
	} else if value < 45 {
		colorFunc = color.YellowString
	} else if value < 55 {
		colorFunc = color.WhiteString
	} else if value < 75 {
		colorFunc = color.GreenString
	} else {
		colorFunc = color.CyanString
	}
	
	fmt.Printf("%s (%s)\n", colorFunc("%d", value), fg.Classification)
	fmt.Printf("   %s\n", fg.Sentiment)
}

func printAnalysisResult(result *types.Analysis, evidenceSummary map[string]interface{}) {
	// Basic info
	fmt.Printf("\n💰 当前价格: %s\n", color.CyanString("$%.2f", result.CurrentPrice))
	fmt.Printf("📈 整体趋势: %s\n", getTrendColor(result.OverallTrend))
	fmt.Printf("📊 趋势评分: %.1f\n", result.TrendScore)

	// Technical indicators table - 展示所有原始数据
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "数值", "参考值", "状态"})
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	// RSI详细信息
	rsiStatus := result.Momentum.Momentum
	rsiRef := "超买>70, 超卖<30"
	table.Append([]string{"RSI(14)", fmt.Sprintf("%.1f", result.Momentum.RSI), rsiRef, rsiStatus})
	
	// MACD详细信息
	table.Append([]string{"MACD", fmt.Sprintf("%.2f", result.MACDAnalysis.MACD), fmt.Sprintf("Signal: %.2f", result.MACDAnalysis.Signal), result.MACDAnalysis.Trend})
	table.Append([]string{"MACD柱", fmt.Sprintf("%.2f", result.MACDAnalysis.Histogram), ">0看涨, <0看跌", ""})
	
	// ADX详细信息
	adxRef := "强势>35, 弱势<20"
	table.Append([]string{"ADX(14)", fmt.Sprintf("%.1f", result.TrendStrength.ADX), adxRef, string(result.TrendStrength.Strength)})
	
	// 成交量详细信息
	volumeRef := "放量>2x, 缩量<0.5x"
	table.Append([]string{"成交量比", fmt.Sprintf("%.2fx", result.Volume.VolumeRatio), volumeRef, result.Volume.VolumeTrend})
	table.Append([]string{"当前成交量", fmt.Sprintf("%.0f", result.Volume.CurrentVolume), fmt.Sprintf("均量: %.0f", result.Volume.VolumeMA), ""})

	fmt.Println("\n📊 技术指标详情:")
	table.Render()

	// Moving averages - 更详细的展示
	fmt.Println("\n📉 移动平均线详情:")
	maTable := tablewriter.NewWriter(os.Stdout)
	maTable.SetHeader([]string{"均线", "价格", "相对位置", "偏离度"})
	maTable.SetBorder(false)
	maTable.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// 计算偏离度
	ma5Deviation := (result.CurrentPrice - result.MAAnalysis.MA5) / result.MAAnalysis.MA5 * 100
	ma20Deviation := (result.CurrentPrice - result.MAAnalysis.MA20) / result.MAAnalysis.MA20 * 100
	ma50Deviation := (result.CurrentPrice - result.MAAnalysis.MA50) / result.MAAnalysis.MA50 * 100
	
	maTable.Append([]string{"MA5", fmt.Sprintf("$%.2f", result.MAAnalysis.MA5), 
		getPriceVsMAIndicator(result.CurrentPrice, result.MAAnalysis.MA5), 
		fmt.Sprintf("%.2f%%", ma5Deviation)})
	maTable.Append([]string{"MA20", fmt.Sprintf("$%.2f", result.MAAnalysis.MA20), 
		getPriceVsMAIndicator(result.CurrentPrice, result.MAAnalysis.MA20), 
		fmt.Sprintf("%.2f%%", ma20Deviation)})
	maTable.Append([]string{"MA50", fmt.Sprintf("$%.2f", result.MAAnalysis.MA50), 
		getPriceVsMAIndicator(result.CurrentPrice, result.MAAnalysis.MA50), 
		fmt.Sprintf("%.2f%%", ma50Deviation)})
	
	maTable.Render()
	
	// 支撑阻力位
	fmt.Println("\n🎯 关键价位:")
	srTable := tablewriter.NewWriter(os.Stdout)
	srTable.SetHeader([]string{"类型", "价位", "距离", "强度"})
	srTable.SetBorder(false)
	srTable.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// 阻力位
	r1Distance := (result.SupportResistance.Resistance["R1"] - result.CurrentPrice) / result.CurrentPrice * 100
	r2Distance := (result.SupportResistance.Resistance["R2"] - result.CurrentPrice) / result.CurrentPrice * 100
	
	srTable.Append([]string{"阻力R2", fmt.Sprintf("$%.2f", result.SupportResistance.Resistance["R2"]), 
		fmt.Sprintf("+%.2f%%", r2Distance), "强"})
	srTable.Append([]string{"阻力R1", fmt.Sprintf("$%.2f", result.SupportResistance.Resistance["R1"]), 
		fmt.Sprintf("+%.2f%%", r1Distance), "中"})
	srTable.Append([]string{"轴心点", fmt.Sprintf("$%.2f", result.SupportResistance.Pivot), 
		"--", "参考"})
	
	// 支撑位
	s1Distance := (result.CurrentPrice - result.SupportResistance.Support["S1"]) / result.CurrentPrice * 100
	s2Distance := (result.CurrentPrice - result.SupportResistance.Support["S2"]) / result.CurrentPrice * 100
	
	srTable.Append([]string{"支撑S1", fmt.Sprintf("$%.2f", result.SupportResistance.Support["S1"]), 
		fmt.Sprintf("-%.2f%%", s1Distance), "中"})
	srTable.Append([]string{"支撑S2", fmt.Sprintf("$%.2f", result.SupportResistance.Support["S2"]), 
		fmt.Sprintf("-%.2f%%", s2Distance), "强"})
	
	srTable.Render()

	// Evidence summary
	bullishCount := evidenceSummary["bullishCount"].(int)
	bearishCount := evidenceSummary["bearishCount"].(int)
	warningCount := evidenceSummary["warningCount"].(int)
	totalStrength := evidenceSummary["totalStrength"].(float64)

	fmt.Printf("\n🔍 证据汇总:\n")
	fmt.Printf("  看涨证据: %s\n", color.GreenString("%d条", bullishCount))
	fmt.Printf("  看跌证据: %s\n", color.RedString("%d条", bearishCount))
	fmt.Printf("  警告信号: %s\n", color.YellowString("%d条", warningCount))
	fmt.Printf("  综合强度: %.2f\n", totalStrength)

	// 详细证据列表
	fmt.Println("\n📋 详细证据分析:")
	evidenceTable := tablewriter.NewWriter(os.Stdout)
	evidenceTable.SetHeader([]string{"类型", "类别", "描述", "权重"})
	evidenceTable.SetBorder(false)
	evidenceTable.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// 显示所有证据
	if allEvidences, ok := evidenceSummary["allEvidences"].([]types.Evidence); ok {
		for _, ev := range allEvidences {
			typeStr := ""
			switch ev.Type {
			case types.BullishEvidence:
				typeStr = "✅ 看涨"
			case types.BearishEvidence:
				typeStr = "❌ 看跌"
			case types.WarningEvidence:
				typeStr = "⚠️ 警告"
			case types.NeutralEvidence:
				typeStr = "➖ 中性"
			}
			evidenceTable.Append([]string{typeStr, ev.Category, ev.Description, fmt.Sprintf("%.2f", ev.Strength)})
		}
	}
	evidenceTable.Render()
	
	// 指标一致性分析
	fmt.Println("\n🔍 指标一致性:")
	fmt.Printf("  看涨信号: %d个\n", bullishCount)
	fmt.Printf("  看跌信号: %d个\n", bearishCount)
	fmt.Printf("  警告信号: %d个\n", warningCount)
	
	consistency := float64(max(bullishCount, bearishCount)) / float64(bullishCount+bearishCount+warningCount) * 100
	fmt.Printf("  一致性: %.1f%%\n", consistency)
	
	// Trading suggestion - 基于原始数据
	fmt.Println("\n💡 参考建议（仅供参考，请结合实际情况）:")
	fmt.Printf("  综合得分: %.2f\n", totalStrength)
	
	if totalStrength > 2 {
		color.Green("  系统判断：强烈看涨信号")
	} else if totalStrength > 0.5 {
		color.Yellow("  系统判断：偏多信号")
	} else if totalStrength < -2 {
		color.Red("  系统判断：强烈看跌信号")
	} else if totalStrength < -0.5 {
		color.Yellow("  系统判断：偏空信号")
	} else {
		fmt.Println("  系统判断：信号不明确")
	}
	
	fmt.Println("\n⚠️  提醒：以上为技术指标分析结果，投资决策需要综合考虑多方面因素")
}

func printPriceChart(ohlcv []types.OHLCV) {
	if len(ohlcv) < 50 {
		return
	}

	// Get last 50 closes
	lastN := 50
	closes := make([]float64, lastN)
	for i := 0; i < lastN; i++ {
		closes[i] = ohlcv[len(ohlcv)-lastN+i].Close
	}

	// Create graph
	graph := asciigraph.Plot(closes, asciigraph.Height(10), asciigraph.Width(60))
	
	fmt.Println("\n📈 价格走势图:")
	fmt.Println(graph)
	
	// Stats
	min, max := closes[0], closes[0]
	for _, v := range closes {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	change := (closes[len(closes)-1] - closes[0]) / closes[0] * 100
	fmt.Printf("\n最高: %.2f  最低: %.2f  变化: %.2f%%\n", max, min, change)
}

func getTrendColor(trend types.TrendDirection) string {
	trendStr := ""
	switch trend {
	case types.StrongUptrend:
		trendStr = "强劲上涨趋势"
		return color.GreenString(trendStr)
	case types.Uptrend:
		trendStr = "上涨趋势"
		return color.GreenString(trendStr)
	case types.Sideways:
		trendStr = "横盘震荡"
		return color.YellowString(trendStr)
	case types.Downtrend:
		trendStr = "下跌趋势"
		return color.RedString(trendStr)
	case types.StrongDowntrend:
		trendStr = "强劲下跌趋势"
		return color.RedString(trendStr)
	default:
		return string(trend)
	}
}

func getPriceVsMAIndicator(price, ma float64) string {
	if price > ma {
		return color.GreenString("↑")
	}
	return color.RedString("↓")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}