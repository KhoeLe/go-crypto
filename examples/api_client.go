package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

// APIResponse represents the standard API response
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Example client demonstrating API usage
func main() {
	fmt.Println("🚀 Go Crypto API Client Example")
	fmt.Println("================================")
	fmt.Println()

	// 1. Health Check
	fmt.Println("1. 🔍 Health Check...")
	healthCheck()

	// 2. Get Price
	fmt.Println("\n2. 💰 Get BTC Price...")
	getPrice("BTCUSDT")

	// 3. Get Indicators
	fmt.Println("\n3. 📊 Get ETH Indicators...")
	getIndicators("ETHUSDT", "15m")

	// 4. Get Analysis
	fmt.Println("\n4. 📈 Get BTC Analysis...")
	getAnalysis("BTCUSDT", "4h")

	// 5. Get Multi-timeframe Analysis
	fmt.Println("\n5. 🕒 Get Multi-timeframe Analysis...")
	getMultiAnalysis("ETHUSDT")

	// 6. Get Signals
	fmt.Println("\n6. 🚨 Get Trading Signals...")
	getSignals("BTCUSDT")

	fmt.Println("\n✅ Demo completed!")
}

func healthCheck() {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ API is healthy!")
	} else {
		fmt.Printf("❌ API health check failed: %d\n", resp.StatusCode)
	}
}

func getPrice(symbol string) {
	url := fmt.Sprintf("%s/price/%s", baseURL, symbol)
	data := makeRequest(url)
	if data != nil {
		var priceData map[string]interface{}
		if err := json.Unmarshal(data, &priceData); err == nil {
			fmt.Printf("💰 %s Price: %v\n", symbol, priceData["price"])
			fmt.Printf("🕒 Updated: %v\n", priceData["timestamp"])
		}
	}
}

func getIndicators(symbol, interval string) {
	url := fmt.Sprintf("%s/indicators/%s?interval=%s", baseURL, symbol, interval)
	data := makeRequest(url)
	if data != nil {
		var indicators map[string]interface{}
		if err := json.Unmarshal(data, &indicators); err == nil {
			fmt.Printf("📊 %s (%s) Indicators:\n", symbol, interval)

			if rsi, ok := indicators["rsi"].(map[string]interface{}); ok {
				fmt.Printf("   RSI Values: %v\n", rsi)
			}

			if ma, ok := indicators["ma"].(map[string]interface{}); ok {
				fmt.Printf("   MA Values: %v\n", ma)
			}

			if kdj, ok := indicators["kdj"].(map[string]interface{}); ok {
				if k, ok := kdj["k"]; ok {
					fmt.Printf("   KDJ K: %v\n", k)
				}
			}
		}
	}
}

func getAnalysis(symbol, interval string) {
	url := fmt.Sprintf("%s/analysis/%s?interval=%s", baseURL, symbol, interval)
	data := makeRequest(url)
	if data != nil {
		var analysis map[string]interface{}
		if err := json.Unmarshal(data, &analysis); err == nil {
			fmt.Printf("📈 %s (%s) Analysis:\n", symbol, interval)

			if price, ok := analysis["price"].(map[string]interface{}); ok {
				fmt.Printf("   Current Price: %v\n", price["price"])
				fmt.Printf("   24h Change: %v%%\n", price["priceChangePercent"])
			}

			if signals, ok := analysis["signals"].([]interface{}); ok {
				fmt.Printf("   🚨 Signals: %v\n", signals)
			}

			if volatility, ok := analysis["volatility"]; ok {
				fmt.Printf("   📊 Volatility: %v\n", volatility)
			}
		}
	}
}

func getMultiAnalysis(symbol string) {
	url := fmt.Sprintf("%s/multi-analysis/%s", baseURL, symbol)
	data := makeRequest(url)
	if data != nil {
		var multiAnalysis map[string]interface{}
		if err := json.Unmarshal(data, &multiAnalysis); err == nil {
			fmt.Printf("🕒 %s Multi-timeframe Analysis:\n", symbol)

			if analyses, ok := multiAnalysis["analyses"].(map[string]interface{}); ok {
				for timeframe, analysis := range analyses {
					fmt.Printf("\n   📊 %s:\n", timeframe)
					if analysisData, ok := analysis.(map[string]interface{}); ok {
						if signals, ok := analysisData["signals"].([]interface{}); ok && len(signals) > 0 {
							fmt.Printf("      🚨 Signals: %v\n", signals)
						} else {
							fmt.Printf("      ✅ No signals\n")
						}
					}
				}
			}
		}
	}
}

func getSignals(symbol string) {
	url := fmt.Sprintf("%s/signals/%s", baseURL, symbol)
	data := makeRequest(url)
	if data != nil {
		var signalData map[string]interface{}
		if err := json.Unmarshal(data, &signalData); err == nil {
			fmt.Printf("🚨 %s Trading Signals:\n", symbol)
			if signals, ok := signalData["signals"].([]interface{}); ok {
				if len(signals) > 0 {
					for _, signal := range signals {
						fmt.Printf("   • %v\n", signal)
					}
				} else {
					fmt.Printf("   ✅ No active signals\n")
				}
			}
		}
	}
}

func makeRequest(url string) json.RawMessage {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Request error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Read error: %v\n", err)
		return nil
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ HTTP %d: %s\n", resp.StatusCode, string(body))
		return nil
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("❌ JSON error: %v\n", err)
		return nil
	}

	if !apiResp.Success {
		fmt.Printf("❌ API error: %s\n", apiResp.Error)
		return nil
	}

	return apiResp.Data
}
