package utils

// CalculateKlineLimit 根据时间间隔和天数计算需要的K线数量
func CalculateKlineLimit(interval string, days int) int {
	switch interval {
	case "15m":
		return days * 24 * 4
	case "30m":
		return days * 24 * 2
	case "1h":
		return days * 24
	case "4h":
		return days * 6
	case "1d":
		return days
	default:
		return days * 24 // default to hourly
	}
}