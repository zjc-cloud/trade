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
	"github.com/zjc/go-crypto-analyzer/pkg/cache"
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
	useCache   bool
	clearCache bool
	cacheDir   string
	cacheTTL   int
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
	rootCmd.Flags().BoolVar(&useCache, "cache", true, "启用数据缓存（默认启用）")
	rootCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "清除所有缓存数据")
	rootCmd.Flags().StringVar(&cacheDir, "cache-dir", ".cache", "缓存目录")
	rootCmd.Flags().IntVar(&cacheTTL, "cache-ttl", 5, "缓存有效期（分钟）")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
	// Handle cache clearing
	if clearCache {
		cacheManager := cache.NewOHLCVCache(cacheDir, time.Duration(cacheTTL)*time.Minute)
		if err := cacheManager.ClearAll(); err != nil {
			color.Red("清除缓存失败: %v", err)
		} else {
			color.Green("✅ 已清除所有缓存数据")
		}
		return
	}

	// Determine symbols to analyze
	symbolsToAnalyze := symbols
	if len(symbolsToAnalyze) == 0 {
		symbolsToAnalyze = config.GetWatchlist(watchlist)
	}

	// Create base data fetcher
	var baseFetcher data.Fetcher
	if useYahoo {
		baseFetcher = data.NewYahooFinanceFetcher()
		fmt.Println("使用Yahoo Finance数据源")
	} else {
		baseFetcher = data.NewBinanceFetcher()
		fmt.Println("使用Binance数据源")
	}

	// Wrap with cache if enabled
	var fetcher data.Fetcher
	if useCache {
		fmt.Printf("✅ 缓存已启用 (目录: %s, TTL: %d分钟)\n", cacheDir, cacheTTL)
		fetcher = data.NewCachedFetcher(baseFetcher, cacheDir, time.Duration(cacheTTL)*time.Minute)
	} else {
		fmt.Println("⚠️  缓存已禁用")
		fetcher = baseFetcher
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

	// 计算实际需要的数据量
	// 1. 技术分析需要至少100根
	// 2. 历史信号追踪需要额外12小时的数据
	minForAnalysis := 100
	extraForHistory := calculatePointsForHours(interval, 12)
	actualLimit := limit
	
	// 如果用户请求的数据不够，自动增加
	minRequired := minForAnalysis + extraForHistory
	if actualLimit < minRequired {
		actualLimit = minRequired
		fmt.Printf("  ℹ️  自动调整数据量: %d → %d (确保历史信号追踪)\n", limit, actualLimit)
	}

	// Fetch OHLCV data
	ohlcv, err := fetcher.FetchOHLCV(symbol, interval, actualLimit)
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
	
	// Print historical signal tracking at the bottom
	printHistoricalSignals(symbol, ohlcv, analyzer, collector)
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

	// Calculate optimal number of data points based on interval
	// Goal: show 5-7 days of data for good trend visibility
	lastN := calculateOptimalDataPoints(interval)
	
	// Ensure we don't exceed available data
	if lastN > len(ohlcv) {
		lastN = len(ohlcv)
	}
	
	// Minimum 50 points for meaningful chart
	if lastN < 50 {
		lastN = 50
	}
	closes := make([]float64, lastN)
	times := make([]time.Time, lastN)
	for i := 0; i < lastN; i++ {
		closes[i] = ohlcv[len(ohlcv)-lastN+i].Close
		times[i] = ohlcv[len(ohlcv)-lastN+i].Time
	}

	// Calculate min and max for price scale
	minPrice, maxPrice := closes[0], closes[0]
	for _, v := range closes {
		if v < minPrice {
			minPrice = v
		}
		if v > maxPrice {
			maxPrice = v
		}
	}
	
	// Create graph with caption
	graph := asciigraph.Plot(closes, 
		asciigraph.Height(10), 
		asciigraph.Width(60),
		asciigraph.Caption(fmt.Sprintf("价格区间: $%.2f - $%.2f", minPrice, maxPrice)))
	
	fmt.Println("\n📈 价格走势图:")
	fmt.Println(graph)
	
	// Time axis
	fmt.Print("    ")
	
	// Format times based on duration
	var startTime, midTime, endTime string
	duration := times[len(times)-1].Sub(times[0])
	
	if duration.Hours() < 24 {
		// Within a day, show hours
		startTime = times[0].Format("15:04")
		endTime = times[len(times)-1].Format("15:04")
		midTime = times[len(times)/2].Format("15:04")
	} else if duration.Hours() < 24*7 {
		// Within a week, show date and hour
		startTime = times[0].Format("01-02 15:04")
		endTime = times[len(times)-1].Format("01-02 15:04")
		midTime = times[len(times)/2].Format("01-02 15:04")
	} else {
		// More than a week, show only date
		startTime = times[0].Format("01-02")
		endTime = times[len(times)-1].Format("01-02")
		midTime = times[len(times)/2].Format("01-02")
	}
	
	// Calculate spacing
	totalWidth := 60
	startLen := len(startTime)
	midLen := len(midTime)
	endLen := len(endTime)
	
	// Print time axis with proper spacing
	fmt.Print(startTime)
	spaces1 := (totalWidth/2 - startLen - midLen/2)
	if spaces1 > 0 {
		fmt.Print(strings.Repeat(" ", spaces1))
	}
	fmt.Print(midTime)
	spaces2 := (totalWidth/2 - midLen/2 - endLen)
	if spaces2 > 0 {
		fmt.Print(strings.Repeat(" ", spaces2))
	}
	fmt.Println(endTime)
	
	// Stats are already calculated above as minPrice and maxPrice
	
	// Time period info
	hoursStr := ""
	if duration.Hours() < 24 {
		hoursStr = fmt.Sprintf("%.0f小时", duration.Hours())
	} else {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		if hours > 0 {
			hoursStr = fmt.Sprintf("%d天%d小时", days, hours)
		} else {
			hoursStr = fmt.Sprintf("%d天", days)
		}
	}
	
	change := (closes[len(closes)-1] - closes[0]) / closes[0] * 100
	fmt.Printf("\n时间跨度: %s  最高: $%.2f  最低: $%.2f  变化: %.2f%%\n", hoursStr, maxPrice, minPrice, change)
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


// calculateOptimalDataPoints 根据时间间隔计算最佳显示点数
func calculateOptimalDataPoints(interval string) int {
	// 平衡图表宽度限制(60字符)和时间跨度
	switch interval {
	case "15m":
		return 80   // 约20小时
	case "30m":
		return 80   // 约40小时  
	case "1h":
		return 120  // 5天
	case "4h":
		return 60   // 10天
	case "1d":
		return 30   // 30天
	default:
		return 80   // 默认值
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// printHistoricalSignals 打印历史信号追踪
func printHistoricalSignals(symbol string, ohlcv []types.OHLCV, analyzer *analysis.TrendAnalyzer, collector *analysis.EvidenceCollector) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 历史信号追踪（过去12小时）")
	fmt.Println(strings.Repeat("=", 80))
	
	// 根据时间间隔计算需要的数据点数
	hoursToShow := 12
	pointsNeeded := calculatePointsForHours(interval, hoursToShow)
	
	// 确保不超过可用数据
	if pointsNeeded > len(ohlcv) {
		pointsNeeded = len(ohlcv)
	}
	
	// 如果数据太少，减少显示的小时数
	if pointsNeeded < 12 {
		hoursToShow = pointsNeeded / calculatePointsPerHour(interval)
		if hoursToShow < 1 {
			fmt.Println("  ⚠️  历史数据不足，无法显示信号追踪")
			return
		}
		fmt.Printf("  ℹ️  数据有限，显示过去%d小时\n", hoursToShow)
	}
	
	// 创建信号追踪表
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "价格", "综合得分", "系统判断", "RSI", "MACD", "成交量"})
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// 存储历史得分用于趋势分析
	var scores []float64
	var times []time.Time
	
	// 从最近的数据开始，向前回溯
	startIdx := len(ohlcv) - pointsNeeded
	minRequired := 100 // 技术分析需要的最小数据点
	
	// 确保有足够的历史数据
	if len(ohlcv) <= minRequired {
		fmt.Println("  ⚠️  数据不足，无法显示完整的历史信号")
		fmt.Printf("  ℹ️  当前只有 %d 根K线，无法同时进行技术分析和历史追踪\n", len(ohlcv))
		return
	}
	
	// 确保startIdx有效
	if startIdx < minRequired {
		startIdx = minRequired
	}
	
	// 确保不会越界
	if startIdx >= len(ohlcv) {
		startIdx = len(ohlcv) - 1
	}
	
	// 计算显示间隔
	totalPoints := len(ohlcv) - startIdx
	if totalPoints <= 0 {
		fmt.Println("  ⚠️  没有足够的历史数据可显示")
		return
	}
	
	maxRows := 24 // 最多显示24行
	step := 1
	if totalPoints > maxRows {
		step = totalPoints / maxRows
		if step < 1 {
			step = 1
		}
	}
	
	// 为了避免重复计算，只在必要时重新分析
	fmt.Printf("\n  ℹ️  分析时间范围: %s 至 %s\n", 
		ohlcv[startIdx].Time.Format("01-02 15:04"),
		ohlcv[len(ohlcv)-1].Time.Format("01-02 15:04"))
	fmt.Printf("  ℹ️  数据点: 共%d个，每%d个显示一次\n\n", totalPoints, step)
	
	for i := startIdx; i < len(ohlcv); i += step {
		// 获取当前时间点的数据窗口（重用已有数据）
		windowStart := i - minRequired + 1
		if windowStart < 0 {
			windowStart = 0
		}
		window := ohlcv[windowStart : i+1]
		
		// 执行技术分析（这里会重用缓存的计算结果）
		result, err := analyzer.AnalyzeComprehensive(window)
		if err != nil {
			continue
		}
		
		// 收集证据
		collector.Clear()
		collector.AnalyzeMAEvidence(result.MAAnalysis, result.CurrentPrice)
		collector.AnalyzeMACDEvidence(result.MACDAnalysis)
		collector.AnalyzeRSIEvidence(result.Momentum.RSI)
		collector.AnalyzeSREvidence(result.CurrentPrice, result.SupportResistance)
		
		// 计算价格变化
		priceChange := 0.0
		if i > 0 {
			priceChange = (window[len(window)-1].Close - window[len(window)-2].Close) / window[len(window)-2].Close
		}
		collector.AnalyzeVolumeEvidence(result.Volume, priceChange)
		
		// 获取综合得分
		summary := collector.GetSummary()
		totalStrength := summary["totalStrength"].(float64)
		
		// 记录数据
		scores = append(scores, totalStrength)
		times = append(times, window[len(window)-1].Time)
		
		// 确定系统判断
		systemJudgment := ""
		if totalStrength > 2 {
			systemJudgment = color.GreenString("强烈看涨信号")
		} else if totalStrength > 0.5 {
			systemJudgment = color.YellowString("偏多信号")
		} else if totalStrength < -2 {
			systemJudgment = color.RedString("强烈看跌信号")
		} else if totalStrength < -0.5 {
			systemJudgment = color.YellowString("偏空信号")
		} else {
			systemJudgment = "信号不明确"
		}
		
		// 格式化MACD
		macdStr := fmt.Sprintf("%.0f", result.MACDAnalysis.MACD)
		if result.MACDAnalysis.MACD > 0 {
			macdStr = color.GreenString(macdStr)
		} else {
			macdStr = color.RedString(macdStr)
		}
		
		// 格式化成交量
		volumeStr := fmt.Sprintf("%.1fx", result.Volume.VolumeRatio)
		if result.Volume.VolumeRatio > 1.5 {
			volumeStr = color.GreenString(volumeStr)
		} else if result.Volume.VolumeRatio < 0.5 {
			volumeStr = color.RedString(volumeStr)
		}
		
		// 添加到表格
		table.Append([]string{
			window[len(window)-1].Time.Format("01-02 15:04"),
			fmt.Sprintf("$%.2f", result.CurrentPrice),
			fmt.Sprintf("%.2f", totalStrength),
			systemJudgment,
			fmt.Sprintf("%.1f", result.Momentum.RSI),
			macdStr,
			volumeStr,
		})
	}
	
	table.Render()
	
	// 分析信号变化趋势
	if len(scores) > 1 {
		fmt.Println("\n🔄 信号变化分析:")
		
		// 计算平均值
		avgScore := 0.0
		for _, s := range scores {
			avgScore += s
		}
		avgScore /= float64(len(scores))
		
		// 找出最高和最低点
		minScore, maxScore := scores[0], scores[0]
		minTime, maxTime := times[0], times[0]
		for i, s := range scores {
			if s < minScore {
				minScore = s
				minTime = times[i]
			}
			if s > maxScore {
				maxScore = s
				maxTime = times[i]
			}
		}
		
		// 趋势判断
		recentAvg := 0.0
		historicalAvg := 0.0
		halfPoint := len(scores) / 2
		
		for i := 0; i < halfPoint; i++ {
			historicalAvg += scores[i]
		}
		historicalAvg /= float64(halfPoint)
		
		for i := halfPoint; i < len(scores); i++ {
			recentAvg += scores[i]
		}
		recentAvg /= float64(len(scores) - halfPoint)
		
		fmt.Printf("  平均得分: %.2f\n", avgScore)
		fmt.Printf("  最高得分: %.2f (%s)\n", maxScore, maxTime.Format("15:04"))
		fmt.Printf("  最低得分: %.2f (%s)\n", minScore, minTime.Format("15:04"))
		
		// 趋势判断
		fmt.Print("  信号趋势: ")
		if recentAvg > historicalAvg + 0.3 {
			color.Green("转强 ↗")
		} else if recentAvg < historicalAvg - 0.3 {
			color.Red("转弱 ↘")
		} else {
			color.Yellow("横盘 →")
		}
		
		// 当前位置
		currentScore := scores[len(scores)-1]
		fmt.Print("\n  当前位置: ")
		if currentScore > avgScore + 1.0 {
			color.Red("可能超买")
		} else if currentScore < avgScore - 1.0 {
			color.Green("可能超卖")
		} else {
			fmt.Println("正常区间")
		}
	}
}

// calculatePointsForHours 根据时间间隔计算需要的数据点数
func calculatePointsForHours(interval string, hours int) int {
	switch interval {
	case "15m":
		return hours * 4
	case "30m":
		return hours * 2
	case "1h":
		return hours
	case "4h":
		return hours / 4
	case "1d":
		return 1
	default:
		return hours
	}
}

// calculatePointsPerHour 计算每小时的数据点数
func calculatePointsPerHour(interval string) int {
	switch interval {
	case "15m":
		return 4
	case "30m":
		return 2
	case "1h":
		return 1
	case "4h":
		return 1 // 4小时返回1，虽然不准确但避免除0
	case "1d":
		return 1 // 1天返回1
	default:
		return 1
	}
}