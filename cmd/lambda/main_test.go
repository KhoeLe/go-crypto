package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestLambdaHandlerProdHealth(t *testing.T) {
	handler, err := NewLambdaHandler()
	if err != nil {
		t.Fatalf("NewLambdaHandler() error = %v", err)
	}

	response, err := handler.HandleRequest(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/prod/api/v1/health",
	})
	if err != nil {
		t.Fatalf("HandleRequest() error = %v", err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200; body=%s", response.StatusCode, response.Body)
	}
	if !strings.Contains(response.Body, "2.1.0-xau-xag-futures") {
		t.Fatalf("health response missing deployment marker: %s", response.Body)
	}
}

func TestLambdaHandlerProdSymbolsIncludesFutures(t *testing.T) {
	handler, err := NewLambdaHandler()
	if err != nil {
		t.Fatalf("NewLambdaHandler() error = %v", err)
	}

	response, err := handler.HandleRequest(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/prod/api/v1/symbols",
	})
	if err != nil {
		t.Fatalf("HandleRequest() error = %v", err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200; body=%s", response.StatusCode, response.Body)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Symbols    []string `json:"symbols"`
			Intervals  []string `json:"intervals"`
			Indicators struct {
				MA struct {
					Periods []int `json:"periods"`
				} `json:"ma"`
			} `json:"indicators"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; body=%s", err, response.Body)
	}
	if !body.Success {
		t.Fatalf("success = false; body=%s", response.Body)
	}

	for _, expected := range []string{"XAUUSDT", "XAGUSDT"} {
		if !containsString(body.Data.Symbols, expected) {
			t.Fatalf("symbols response missing %q: %s", expected, response.Body)
		}
	}
	if !containsString(body.Data.Intervals, "1h") {
		t.Fatalf("symbols response missing 1h interval: %s", response.Body)
	}
	if !containsInt(body.Data.Indicators.MA.Periods, 200) {
		t.Fatalf("symbols response missing MA200 period: %s", response.Body)
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsInt(values []int, expected int) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
