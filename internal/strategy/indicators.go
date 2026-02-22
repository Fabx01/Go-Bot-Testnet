package strategy

import "fmt"

func emaSeries(prices []float64, period int) ([]float64, error) {
	if period <= 1 {
		return nil, fmt.Errorf("period must be > 1")
	}
	if len(prices) < period {
		return nil, fmt.Errorf("insufficient data for EMA")
	}

	series := make([]float64, len(prices))
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	series[period-1] = sum / float64(period)

	multiplier := 2.0 / (float64(period) + 1.0)
	for i := period; i < len(prices); i++ {
		series[i] = ((prices[i] - series[i-1]) * multiplier) + series[i-1]
	}
	return series, nil
}

func rsi(prices []float64, period int) (float64, error) {
	if period <= 1 {
		return 0, fmt.Errorf("period must be > 1")
	}
	if len(prices) < period+1 {
		return 0, fmt.Errorf("insufficient data for RSI")
	}

	gain := 0.0
	loss := 0.0
	for i := 1; i <= period; i++ {
		delta := prices[i] - prices[i-1]
		if delta > 0 {
			gain += delta
		} else {
			loss -= delta
		}
	}

	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)

	for i := period + 1; i < len(prices); i++ {
		delta := prices[i] - prices[i-1]
		currentGain := 0.0
		currentLoss := 0.0
		if delta > 0 {
			currentGain = delta
		} else {
			currentLoss = -delta
		}
		avgGain = ((avgGain * float64(period-1)) + currentGain) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + currentLoss) / float64(period)
	}

	if avgLoss == 0 {
		return 100, nil
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs)), nil
}
