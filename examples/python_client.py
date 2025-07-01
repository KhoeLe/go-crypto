#!/usr/bin/env python3
"""
Go Crypto API - Python Client Example

This example shows how to use the Go Crypto Trading Analysis API
from Python applications.

Requirements:
    pip install requests

Usage:
    python examples/python_client.py
"""

import requests
import json
import time
from datetime import datetime

class GoCryptoAPI:
    """Python client for Go Crypto Trading Analysis API"""
    
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.api_base = f"{base_url}/api/v1"
    
    def _make_request(self, endpoint, params=None):
        """Make HTTP request to API"""
        try:
            url = f"{self.api_base}{endpoint}"
            response = requests.get(url, params=params, timeout=10)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"API request failed: {e}")
            return None
    
    def health_check(self):
        """Check API health"""
        return self._make_request("/health")
    
    def get_price(self, symbol):
        """Get current price for symbol"""
        return self._make_request(f"/price/{symbol}")
    
    def get_ticker(self, symbol):
        """Get 24h ticker data"""
        return self._make_request(f"/ticker/{symbol}")
    
    def get_analysis(self, symbol, interval="15m", limit=50):
        """Get complete technical analysis"""
        params = {"interval": interval, "limit": limit}
        return self._make_request(f"/analysis/{symbol}", params)
    
    def get_signals(self, symbol, interval="15m"):
        """Get trading signals"""
        params = {"interval": interval}
        return self._make_request(f"/signals/{symbol}", params)
    
    def get_multi_analysis(self, symbol):
        """Get multi-timeframe analysis"""
        return self._make_request(f"/multi-analysis/{symbol}")
    
    def get_symbols(self):
        """Get available symbols and intervals"""
        return self._make_request("/symbols")

def print_analysis(analysis_data):
    """Pretty print analysis data"""
    if not analysis_data or not analysis_data.get('success'):
        print("❌ Failed to get analysis data")
        return
    
    data = analysis_data['data']
    
    print(f"\n📊 Analysis for {data['symbol']} ({data['timeframe']})")
    print("=" * 50)
    
    # Price info
    price = data['price']
    print(f"💰 Current Price: ${price['price']}")
    print(f"📈 24h Change: {price['priceChangePercent']}%")
    print(f"📊 Volume: {price['volume']}")
    
    # Technical indicators
    print(f"\n🔍 Technical Indicators:")
    rsi = data['rsi']
    for period, value in rsi.items():
        print(f"   RSI-{period.split('_')[1]}: {float(value):.2f}")
    
    ma = data['ma']
    for period, value in ma.items():
        print(f"   MA-{period.split('_')[1]}: ${float(value):.2f}")
    
    kdj = data['kdj']
    print(f"   KDJ K: {float(kdj['k']):.2f}")
    print(f"   KDJ D: {float(kdj['d']):.2f}")
    print(f"   KDJ J: {float(kdj['j']):.2f}")
    
    # Signals
    if data['signals']:
        print(f"\n🚨 Trading Signals:")
        for signal in data['signals']:
            print(f"   • {signal}")
    else:
        print(f"\n✅ No active signals")

def main():
    """Main example function"""
    print("🚀 Go Crypto API - Python Client Example")
    print("=" * 50)
    
    # Initialize API client
    api = GoCryptoAPI()
    
    # Health check
    print("🔍 Checking API health...")
    health = api.health_check()
    if health and health.get('success'):
        print("✅ API is healthy!")
    else:
        print("❌ API is not responding. Make sure the server is running.")
        print("Start the server with: make start-api")
        return
    
    # Get available symbols
    print("\n📋 Getting available symbols...")
    symbols_data = api.get_symbols()
    if symbols_data and symbols_data.get('success'):
        symbols = symbols_data['data']['symbols'][:3]  # Show first 3
        print(f"Available symbols: {', '.join(symbols)}")
    
    # Analyze multiple symbols
    test_symbols = ['BTCUSDT', 'ETHUSDT']
    
    for symbol in test_symbols:
        print(f"\n🔎 Analyzing {symbol}...")
        
        # Get current price
        price_data = api.get_price(symbol)
        if price_data and price_data.get('success'):
            price = price_data['data']['price']
            print(f"💲 {symbol} Price: ${price}")
        
        # Get complete analysis
        analysis = api.get_analysis(symbol, interval="15m")
        print_analysis(analysis)
        
        # Get trading signals
        signals_data = api.get_signals(symbol)
        if signals_data and signals_data.get('success'):
            signals = signals_data['data']['signals']
            if signals:
                print(f"⚡ Active Signals: {', '.join(signals)}")
    
    # Multi-timeframe example
    print(f"\n🕐 Multi-timeframe analysis for BTCUSDT...")
    multi_data = api.get_multi_analysis('BTCUSDT')
    if multi_data and multi_data.get('success'):
        timeframes = multi_data['data']['timeframes']
        print(f"📊 Analyzed timeframes: {', '.join(timeframes.keys())}")
        
        for tf, data in timeframes.items():
            signals = data.get('signals', [])
            rsi_12 = float(data['rsi']['RSI_12'])
            print(f"   {tf}: RSI={rsi_12:.1f}, Signals={len(signals)}")

if __name__ == "__main__":
    main()
