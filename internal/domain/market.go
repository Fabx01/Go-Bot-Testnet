package domain

import "time"

type Candle struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
}

type Signal string

const (
	SignalHold Signal = "HOLD"
	SignalBuy  Signal = "BUY"
	SignalSell Signal = "SELL"
)

type Position struct {
	EntryPrice float64
	BaseQty    float64
	OpenedAt   time.Time
}

type AccountBalance struct {
	Asset string
	Free  float64
}

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderResult struct {
	Symbol      string
	Side        OrderSide
	ExecutedQty float64
	Price       float64
	OrderID     int64
}
