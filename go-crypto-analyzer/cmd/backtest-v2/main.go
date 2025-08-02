package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/zjc/go-crypto-analyzer/pkg/backtest"
	"github.com/zjc/go-crypto-analyzer/pkg/data"
)

var (
	symbol         string
	interval       string
	days           int
	initialCapital float64
	longThreshold  float64
	shortThreshold float64
	closeThreshold float64
	stopLoss       float64
	takeProfit     float64
	useYahoo       bool
	enableShort    bool
	useImproved    bool
)

var rootCmd = &cobra.Command{
	Use:   "backtest-v2",
	Short: "双向交易策略回测",
	Long:  `支持做多做空的交易策略回测工具`,
	Run:   runBacktest,
}

func init() {
	rootCmd.Flags().StringVarP(&symbol, "symbol", "s", "BTCUSDT", "交易对")
	rootCmd.Flags().StringVarP(&interval, "interval", "i", "1h", "K线时间间隔")
	rootCmd.Flags().IntVarP(&days, "days", "d", 30, "回测天数")
	rootCmd.Flags().Float64VarP(&initialCapital, "capital", "c", 10000, "初始资金")
	rootCmd.Flags().Float64VarP(&longThreshold, "long", "L", 0.5, "做多阈值")
	rootCmd.Flags().Float64VarP(&shortThreshold, "short", "S", -0.5, "做空阈值")
	rootCmd.Flags().Float64VarP(&closeThreshold, "close", "C", 0.0, "平仓阈值")
	rootCmd.Flags().Float64VarP(&stopLoss, "stoploss", "l", 0.03, "止损百分比")
	rootCmd.Flags().Float64VarP(&takeProfit, "takeprofit", "t", 0.06, "止盈百分比")
	rootCmd.Flags().BoolVarP(&useYahoo, "yahoo", "y", false, "使用Yahoo Finance数据源")
	rootCmd.Flags().BoolVarP(&enableShort, "enable-short", "E", true, "启用做空")
	rootCmd.Flags().BoolVarP(&useImproved, "improved", "I", false, "使用改进的策略")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBacktest(cmd *cobra.Command, args []string) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("📊 双向交易回测 - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	
	// 创建数据获取器
	var fetcher data.Fetcher
	if useYahoo {
		fetcher = data.NewYahooFinanceFetcher()
		fmt.Println("使用Yahoo Finance数据源")
	} else {
		fetcher = data.NewBinanceFetcher()
		fmt.Println("使用Binance数据源")
	}
	
	// 计算需要的K线数量
	limit := calculateLimit(interval, days)
	
	fmt.Printf("\n⏳ 获取历史数据: %s, %s, %d根K线...\n", symbol, interval, limit)
	
	// 获取历史数据
	ohlcv, err := fetcher.FetchOHLCV(symbol, interval, limit)
	if err != nil {
		color.Red("❌ 获取数据失败: %v", err)
		return
	}
	
	fmt.Printf("✅ 成功获取 %d 根K线数据\n", len(ohlcv))
	
	// 创建支持做空的回测器
	backtester := backtest.NewBacktesterV2(initialCapital)
	backtester.EnableShort(enableShort)
	backtester.SetThresholds(longThreshold, shortThreshold, closeThreshold)
	backtester.UseImprovedStrategy(useImproved)
	
	fmt.Printf("\n📈 回测参数:\n")
	fmt.Printf("  初始资金: $%.2f\n", initialCapital)
	if !useImproved {
		fmt.Printf("  做多阈值: %.2f\n", longThreshold)
		fmt.Printf("  做空阈值: %.2f\n", shortThreshold)
		fmt.Printf("  平仓阈值: %.2f\n", closeThreshold)
		fmt.Printf("  止损: %.1f%%\n", stopLoss*100)
		fmt.Printf("  止盈: %.1f%%\n", takeProfit*100)
	}
	if enableShort {
		color.Green("  ✅ 启用做空")
	} else {
		color.Yellow("  ❌ 禁用做空")
	}
	if useImproved {
		color.Cyan("  📊 使用改进的策略（动态止损、市场状态适应）")
	} else {
		fmt.Printf("  📊 使用基础策略")
	}
	
	fmt.Printf("\n⚙️  运行回测...\n")
	
	// 运行回测
	result, err := backtester.RunBacktestV2(symbol, ohlcv)
	if err != nil {
		color.Red("❌ 回测失败: %v", err)
		return
	}
	
	// 显示结果
	displayResults(result)
}

func calculateLimit(interval string, days int) int {
	switch interval {
	case "15m":
		return days * 24 * 4 + 100
	case "30m":
		return days * 24 * 2 + 100
	case "1h":
		return days * 24 + 100
	case "4h":
		return days * 6 + 100
	case "1d":
		return days + 100
	default:
		return days * 24 + 100
	}
}

func displayResults(result *backtest.BacktestResultV2) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("📊 双向交易回测结果")
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	
	// 基本统计
	fmt.Printf("\n💰 资金变化:\n")
	fmt.Printf("  初始资金: $%.2f\n", result.InitialCapital)
	fmt.Printf("  最终资金: ")
	if result.FinalCapital > result.InitialCapital {
		color.Green("$%.2f", result.FinalCapital)
	} else {
		color.Red("$%.2f", result.FinalCapital)
	}
	fmt.Printf("\n  总收益: ")
	if result.TotalReturn > 0 {
		color.Green("$%.2f (%.2f%%)", result.TotalReturn, result.TotalReturnPct*100)
	} else {
		color.Red("$%.2f (%.2f%%)", result.TotalReturn, result.TotalReturnPct*100)
	}
	fmt.Printf("\n  最大回撤: ")
	color.Red("%.2f%%", result.MaxDrawdownPct*100)
	
	fmt.Printf("\n\n📈 交易统计:\n")
	fmt.Printf("  总交易次数: %d\n", result.TotalTrades)
	fmt.Printf("  做多交易: %s\n", color.BlueString("%d", result.LongTrades))
	fmt.Printf("  做空交易: %s\n", color.MagentaString("%d", result.ShortTrades))
	fmt.Printf("  获利交易: %s\n", color.GreenString("%d", result.WinningTrades))
	fmt.Printf("  亏损交易: %s\n", color.RedString("%d", result.LosingTrades))
	fmt.Printf("  胜率: %.1f%%\n", result.WinRate*100)
	
	if result.AverageWin > 0 || result.AverageLoss > 0 {
		fmt.Printf("\n💵 盈亏分析:\n")
		fmt.Printf("  平均盈利: $%.2f\n", result.AverageWin)
		fmt.Printf("  平均亏损: $%.2f\n", result.AverageLoss)
		if result.ProfitFactor > 0 {
			fmt.Printf("  盈亏比: %.2f\n", result.ProfitFactor)
		}
	}
	
	fmt.Printf("\n📊 风险指标:\n")
	fmt.Printf("  夏普比率: %.2f\n", result.SharpeRatio)
	fmt.Printf("  卡尔玛比率: %.2f\n", result.CalmarRatio)
	
	// 做多做空统计
	if result.LongTrades > 0 || result.ShortTrades > 0 {
		fmt.Printf("\n📊 方向统计:\n")
		
		longWins := 0
		shortWins := 0
		longProfit := 0.0
		shortProfit := 0.0
		
		for _, trade := range result.Trades {
			if trade.Direction == "LONG" {
				if trade.Profit > 0 {
					longWins++
				}
				longProfit += trade.Profit
			} else if trade.Direction == "SHORT" {
				if trade.Profit > 0 {
					shortWins++
				}
				shortProfit += trade.Profit
			}
		}
		
		if result.LongTrades > 0 {
			longWinRate := float64(longWins) / float64(result.LongTrades) * 100
			fmt.Printf("  做多胜率: %.1f%% (盈亏: ", longWinRate)
			if longProfit > 0 {
				color.Green("$%.2f", longProfit)
			} else {
				color.Red("$%.2f", longProfit)
			}
			fmt.Printf(")\n")
		}
		
		if result.ShortTrades > 0 {
			shortWinRate := float64(shortWins) / float64(result.ShortTrades) * 100
			fmt.Printf("  做空胜率: %.1f%% (盈亏: ", shortWinRate)
			if shortProfit > 0 {
				color.Green("$%.2f", shortProfit)
			} else {
				color.Red("$%.2f", shortProfit)
			}
			fmt.Printf(")\n")
		}
	}
	
	// 交易明细表
	if len(result.Trades) > 0 {
		fmt.Printf("\n📋 交易明细 (最近15笔):\n")
		
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"时间", "方向", "入场价", "出场价", "收益", "收益率", "信号"})
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		
		// 显示最近15笔交易
		start := 0
		if len(result.Trades) > 15 {
			start = len(result.Trades) - 15
		}
		
		for i := start; i < len(result.Trades); i++ {
			trade := result.Trades[i]
			
			directionStr := trade.Direction
			if trade.Direction == "LONG" {
				directionStr = color.BlueString("做多")
			} else if trade.Direction == "SHORT" {
				directionStr = color.MagentaString("做空")
			}
			
			profitStr := fmt.Sprintf("$%.2f", trade.Profit)
			profitPctStr := fmt.Sprintf("%.2f%%", trade.ProfitPct*100)
			
			if trade.Profit > 0 {
				profitStr = color.GreenString(profitStr)
				profitPctStr = color.GreenString(profitPctStr)
			} else {
				profitStr = color.RedString(profitStr)
				profitPctStr = color.RedString(profitPctStr)
			}
			
			table.Append([]string{
				trade.EntryTime.Format("01-02 15:04"),
				directionStr,
				fmt.Sprintf("$%.2f", trade.EntryPrice),
				fmt.Sprintf("$%.2f", trade.ExitPrice),
				profitStr,
				profitPctStr,
				trade.ExitSignal,
			})
		}
		
		table.Render()
		
		fmt.Printf("\n共 %d 笔交易，显示最近 %d 笔\n", len(result.Trades), len(result.Trades)-start)
	}
	
	// 策略评价
	fmt.Printf("\n💡 策略评价:\n")
	if result.TotalReturnPct > 0.2 {
		color.Green("  ✅ 策略表现优秀，年化收益可观")
	} else if result.TotalReturnPct > 0 {
		color.Yellow("  ⚠️  策略有盈利，但收益率一般")
	} else {
		color.Red("  ❌ 策略亏损，需要优化参数或改进策略")
	}
	
	if result.MaxDrawdownPct > 0.2 {
		color.Red("  ⚠️  最大回撤较大，风险控制需要加强")
	}
	
	if result.WinRate < 0.4 {
		color.Yellow("  ⚠️  胜率较低，考虑优化入场条件")
	}
	
	if result.SharpeRatio < 1 {
		color.Yellow("  ⚠️  夏普比率较低，收益风险比需要改善")
	} else if result.SharpeRatio > 2 {
		color.Green("  ✅ 夏普比率优秀，风险调整后收益良好")
	}
	
	if result.CalmarRatio > 3 {
		color.Green("  ✅ 卡尔玛比率优秀，收益回撤比良好")
	} else if result.CalmarRatio < 1 {
		color.Red("  ⚠️  卡尔玛比率较低，回撤控制需要改善")
	}
}