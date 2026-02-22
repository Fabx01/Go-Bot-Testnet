package strategy

import "testing"

func TestEMASeriesAndRSI(t *testing.T) {
	prices := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115}

	ema, err := emaSeries(prices, 5)
	if err != nil {
		t.Fatalf("emaSeries returned error: %v", err)
	}
	if len(ema) != len(prices) {
		t.Fatalf("unexpected ema length: got %d want %d", len(ema), len(prices))
	}
	if ema[len(ema)-1] <= ema[4] {
		t.Fatalf("expected ema to trend upward")
	}

	rsiValue, err := rsi(prices, 14)
	if err != nil {
		t.Fatalf("rsi returned error: %v", err)
	}
	if rsiValue <= 50 {
		t.Fatalf("expected RSI > 50 for strong uptrend, got %.2f", rsiValue)
	}
}
