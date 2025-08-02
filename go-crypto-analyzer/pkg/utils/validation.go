package utils

import (
	"fmt"
	"strings"
)

// ValidateSymbol 验证交易对格式
func ValidateSymbol(symbol string) error {
	symbol = strings.ToUpper(symbol)
	
	// 检查基本格式
	if len(symbol) < 6 {
		return fmt.Errorf("invalid symbol format: %s", symbol)
	}
	
	// 检查是否包含USDT
	if !strings.HasSuffix(symbol, "USDT") && !strings.Contains(symbol, "-USD") {
		return fmt.Errorf("symbol must be a USDT pair: %s", symbol)
	}
	
	return nil
}

// ValidateInterval 验证时间间隔
func ValidateInterval(interval string) error {
	validIntervals := map[string]bool{
		"1m": true, "3m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "2h": true, "4h": true, "6h": true, "8h": true, "12h": true,
		"1d": true, "3d": true, "1w": true, "1M": true,
	}
	
	if !validIntervals[interval] {
		return fmt.Errorf("invalid interval: %s", interval)
	}
	
	return nil
}

// ValidateDataQuality 验证数据质量
func ValidateDataQuality(closes []float64) error {
	if len(closes) == 0 {
		return fmt.Errorf("no data available")
	}
	
	// 检查数据中的异常值
	zeroCount := 0
	for _, price := range closes {
		if price <= 0 {
			zeroCount++
		}
	}
	
	if zeroCount > len(closes)/10 {
		return fmt.Errorf("too many invalid prices (zeros): %d/%d", zeroCount, len(closes))
	}
	
	// 检查价格是否合理
	maxPrice := closes[0]
	minPrice := closes[0]
	for _, price := range closes {
		if price > maxPrice {
			maxPrice = price
		}
		if price < minPrice && price > 0 {
			minPrice = price
		}
	}
	
	// 如果最高价是最低价的100倍以上，可能有问题
	if maxPrice > minPrice*100 {
		return fmt.Errorf("price range too large: min=%.2f, max=%.2f", minPrice, maxPrice)
	}
	
	return nil
}