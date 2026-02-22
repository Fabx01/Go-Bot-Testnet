package observability

import (
	"time"

	"github.com/seuprojeto/tradergobinance/internal/domain"
)

type Recorder interface {
	ObserveCycle(duration time.Duration, err error)
	ObserveSignal(signal domain.Signal)
	SetLastPrice(price float64)
	SetPosition(entryPrice, quantity, currentPrice float64)
	ClearPosition()
	ObserveTrade(side domain.OrderSide, mode string, pnlPct, pnlQuote float64)
}

type NoopRecorder struct{}

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (n *NoopRecorder) ObserveCycle(_ time.Duration, _ error) {}

func (n *NoopRecorder) ObserveSignal(_ domain.Signal) {}

func (n *NoopRecorder) SetLastPrice(_ float64) {}

func (n *NoopRecorder) SetPosition(_, _, _ float64) {}

func (n *NoopRecorder) ClearPosition() {}

func (n *NoopRecorder) ObserveTrade(_ domain.OrderSide, _ string, _, _ float64) {}
