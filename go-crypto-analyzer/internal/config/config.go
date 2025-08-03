package config

import (
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// CryptoConfig holds configuration for different cryptocurrencies
var CryptoConfig = map[string]types.CryptoConfig{
	"BTCUSDT": {
		Symbol:     "BTCUSDT",
		Name:       "比特币",
		Category:   "store_of_value",
		Weight:     0.5,
		Timeframes: []string{"15m", "1h", "4h", "1d"},
		KeyLevels: types.KeyLevels{
			Psychological:        []float64{100000, 110000, 120000, 150000},
			HistoricalSupport:    []float64{92000, 85000, 80000},
			HistoricalResistance: []float64{120000, 125000, 130000},
		},
	},
	"ETHUSDT": {
		Symbol:     "ETHUSDT",
		Name:       "以太坊",
		Category:   "smart_contract",
		Weight:     0.3,
		Timeframes: []string{"15m", "1h", "4h", "1d"},
		KeyLevels: types.KeyLevels{
			Psychological:        []float64{3000, 3500, 4000, 5000},
			HistoricalSupport:    []float64{2800, 2500, 2200},
			HistoricalResistance: []float64{3500, 4000, 4500},
		},
	},
	"BNBUSDT": {
		Symbol:     "BNBUSDT",
		Name:       "币安币",
		Category:   "exchange",
		Weight:     0.1,
		Timeframes: []string{"15m", "1h", "4h", "1d"},
		KeyLevels: types.KeyLevels{
			Psychological:        []float64{500, 600, 700, 800},
			HistoricalSupport:    []float64{450, 400, 350},
			HistoricalResistance: []float64{650, 700, 750},
		},
	},
}

// Watchlists defines preset watchlists
var Watchlists = map[string][]string{
	"top3":   {"BTCUSDT", "ETHUSDT", "BNBUSDT"},
	"top10":  {"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "ADAUSDT", "AVAXUSDT", "DOTUSDT", "MATICUSDT", "LINKUSDT", "NEARUSDT"},
	"defi":   {"UNIUSDT", "AAVEUSDT", "LINKUSDT", "MKRUSDT"},
	"layer1": {"ETHUSDT", "SOLUSDT", "AVAXUSDT", "ADAUSDT", "DOTUSDT"},
}

// GetWatchlist returns a watchlist by name
func GetWatchlist(name string) []string {
	if list, ok := Watchlists[name]; ok {
		return list
	}
	return Watchlists["top3"]
}