package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/zjc/go-crypto-analyzer/pkg/types"
)

// Fetcher interface defines methods for fetching market data
type Fetcher interface {
	FetchOHLCV(symbol string, interval string, limit int) ([]types.OHLCV, error)
}

// BinanceFetcher implements Fetcher for Binance exchange
type BinanceFetcher struct {
	client *binance.Client
}

// NewBinanceFetcher creates a new BinanceFetcher
func NewBinanceFetcher() *BinanceFetcher {
	client := binance.NewClient("", "")
	return &BinanceFetcher{client: client}
}

// FetchOHLCV fetches OHLCV data from Binance
func (bf *BinanceFetcher) FetchOHLCV(symbol string, interval string, limit int) ([]types.OHLCV, error) {
	klines, err := bf.client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}

	data := make([]types.OHLCV, len(klines))
	for i, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		close, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		data[i] = types.OHLCV{
			Time:   time.Unix(k.OpenTime/1000, 0),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		}
	}

	return data, nil
}


// YahooFinanceFetcher implements Fetcher for Yahoo Finance
type YahooFinanceFetcher struct {
	client *http.Client
}

// NewYahooFinanceFetcher creates a new YahooFinanceFetcher
func NewYahooFinanceFetcher() *YahooFinanceFetcher {
	return &YahooFinanceFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// symbolMapping maps crypto symbols to Yahoo Finance symbols
var symbolMapping = map[string]string{
	"BTCUSDT": "BTC-USD",
	"ETHUSDT": "ETH-USD",
	"BNBUSDT": "BNB-USD",
	"SOLUSDT": "SOL-USD",
	"ADAUSDT": "ADA-USD",
}

// FetchOHLCV fetches OHLCV data from Yahoo Finance
func (yf *YahooFinanceFetcher) FetchOHLCV(symbol string, interval string, limit int) ([]types.OHLCV, error) {
	// Map symbol
	yfSymbol, ok := symbolMapping[symbol]
	if !ok {
		yfSymbol = symbol
	}

	// Calculate time range
	endTime := time.Now().Unix()
	startTime := endTime - int64(limit*86400) // Approximate based on daily candles

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=%s",
		yfSymbol, startTime, endTime, yf.mapInterval(interval),
	)

	resp, err := yf.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result YahooResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned")
	}

	chartData := result.Chart.Result[0]
	timestamps := chartData.Timestamp
	quote := chartData.Indicators.Quote[0]

	data := make([]types.OHLCV, len(timestamps))
	for i := range timestamps {
		data[i] = types.OHLCV{
			Time:   time.Unix(timestamps[i], 0),
			Open:   quote.Open[i],
			High:   quote.High[i],
			Low:    quote.Low[i],
			Close:  quote.Close[i],
			Volume: quote.Volume[i],
		}
	}

	return data, nil
}


// mapInterval maps standard intervals to Yahoo Finance intervals
func (yf *YahooFinanceFetcher) mapInterval(interval string) string {
	mapping := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"60m": "60m",
		"1h":  "60m",
		"4h":  "60m", // Yahoo doesn't support 4h, use 60m
		"1d":  "1d",
		"1w":  "1wk",
	}

	if mapped, ok := mapping[interval]; ok {
		return mapped
	}
	return "1d"
}

// YahooResponse represents Yahoo Finance API response structure
type YahooResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// FearGreedFetcher fetches the Fear and Greed Index
type FearGreedFetcher struct {
	client *http.Client
}

// NewFearGreedFetcher creates a new FearGreedFetcher
func NewFearGreedFetcher() *FearGreedFetcher {
	return &FearGreedFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Fetch gets the current Fear and Greed Index
func (fg *FearGreedFetcher) Fetch() (*types.FearGreedIndex, error) {
	resp, err := fg.client.Get("https://api.alternative.me/fng/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fear greed index: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Value               string `json:"value"`
			ValueClassification string `json:"value_classification"`
			Timestamp           string `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no data available")
	}

	value, _ := strconv.Atoi(result.Data[0].Value)
	timestamp, _ := strconv.ParseInt(result.Data[0].Timestamp, 10, 64)

	sentiment := ""
	if value < 25 {
		sentiment = "极度恐慌 - 可能是买入机会"
	} else if value < 45 {
		sentiment = "恐慌 - 市场偏空"
	} else if value < 55 {
		sentiment = "中性 - 观望为主"
	} else if value < 75 {
		sentiment = "贪婪 - 市场偏多"
	} else {
		sentiment = "极度贪婪 - 注意风险"
	}

	return &types.FearGreedIndex{
		Value:          value,
		Classification: result.Data[0].ValueClassification,
		Sentiment:      sentiment,
		Timestamp:      time.Unix(timestamp, 0),
	}, nil
}