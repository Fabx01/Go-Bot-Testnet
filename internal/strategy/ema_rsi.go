package strategy

import (
	"fmt"

	"github.com/seuprojeto/tradergobinance/internal/domain"
)

type EMARSI struct {
	ShortPeriod     int
	LongPeriod      int
	RSIPeriod       int
	OverboughtLevel float64
}

func NewEMARSI(shortPeriod, longPeriod, rsiPeriod int, overbought float64) *EMARSI {
	return &EMARSI{
		ShortPeriod:     shortPeriod,
		LongPeriod:      longPeriod,
		RSIPeriod:       rsiPeriod,
		OverboughtLevel: overbought,
	}
}

func (s *EMARSI) Name() string {
	return "ema_rsi"
}

func (s *EMARSI) Evaluate(candles []domain.Candle, position *domain.Position) (domain.Signal, string) {
	minCandles := s.LongPeriod + 2
	if s.RSIPeriod+2 > minCandles {
		minCandles = s.RSIPeriod + 2
	}
	if len(candles) < minCandles {
		return domain.SignalHold, "candles insuficientes"
	}

	closes := make([]float64, 0, len(candles))
	for _, candle := range candles {
		closes = append(closes, candle.Close)
	}

	shortEMA, err := emaSeries(closes, s.ShortPeriod)
	if err != nil {
		return domain.SignalHold, err.Error()
	}
	longEMA, err := emaSeries(closes, s.LongPeriod)
	if err != nil {
		return domain.SignalHold, err.Error()
	}
	lastRSI, err := rsi(closes, s.RSIPeriod)
	if err != nil {
		return domain.SignalHold, err.Error()
	}

	lastIndex := len(closes) - 1
	prevIndex := len(closes) - 2

	prevCrossBelow := shortEMA[prevIndex] <= longEMA[prevIndex]
	currCrossAbove := shortEMA[lastIndex] > longEMA[lastIndex]
	currCrossBelow := shortEMA[lastIndex] < longEMA[lastIndex]

	if position == nil && prevCrossBelow && currCrossAbove && lastRSI < s.OverboughtLevel {
		return domain.SignalBuy, fmt.Sprintf("EMA cross up e RSI=%.2f", lastRSI)
	}

	if position != nil && (currCrossBelow || lastRSI >= s.OverboughtLevel) {
		return domain.SignalSell, fmt.Sprintf("sinal de saida (EMA/RSI=%.2f)", lastRSI)
	}

	return domain.SignalHold, fmt.Sprintf("sem setup (RSI=%.2f)", lastRSI)
}
