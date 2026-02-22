package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/seuprojeto/tradergobinance/internal/domain"
)

type PrometheusRecorder struct {
	cyclesTotal          *prometheus.CounterVec
	cycleDuration        prometheus.Histogram
	signalsTotal         *prometheus.CounterVec
	tradesTotal          *prometheus.CounterVec
	positionOpen         prometheus.Gauge
	lastPrice            prometheus.Gauge
	unrealizedPnlPct     prometheus.Gauge
	lastTradePnlPct      prometheus.Gauge
	lastTradePnlQuote    prometheus.Gauge
	cumulativePnlQuote   prometheus.Gauge
	cumulativePnlPercent prometheus.Gauge
}

func NewPrometheusRecorder(registry *prometheus.Registry) *PrometheusRecorder {
	r := &PrometheusRecorder{
		cyclesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "traderbot_cycles_total",
			Help: "Total de ciclos executados pelo engine.",
		}, []string{"status"}),
		cycleDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "traderbot_cycle_duration_seconds",
			Help:    "Duracao do ciclo do engine em segundos.",
			Buckets: prometheus.DefBuckets,
		}),
		signalsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "traderbot_signals_total",
			Help: "Quantidade de sinais por tipo.",
		}, []string{"signal"}),
		tradesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "traderbot_trades_total",
			Help: "Quantidade de trades executados por lado/modo.",
		}, []string{"side", "mode"}),
		positionOpen: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_position_open",
			Help: "Indica se ha posicao aberta (1) ou nao (0).",
		}),
		lastPrice: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_last_price",
			Help: "Ultimo preco de mercado processado.",
		}),
		unrealizedPnlPct: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_unrealized_pnl_percent",
			Help: "PnL percentual nao realizado da posicao atual.",
		}),
		lastTradePnlPct: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_last_trade_pnl_percent",
			Help: "PnL percentual do ultimo trade de saida.",
		}),
		lastTradePnlQuote: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_last_trade_pnl_quote",
			Help: "PnL em moeda de cotacao do ultimo trade de saida.",
		}),
		cumulativePnlQuote: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_cumulative_pnl_quote",
			Help: "PnL acumulado em moeda de cotacao.",
		}),
		cumulativePnlPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "traderbot_cumulative_pnl_percent",
			Help: "Soma simples do PnL percentual acumulado.",
		}),
	}

	registry.MustRegister(
		r.cyclesTotal,
		r.cycleDuration,
		r.signalsTotal,
		r.tradesTotal,
		r.positionOpen,
		r.lastPrice,
		r.unrealizedPnlPct,
		r.lastTradePnlPct,
		r.lastTradePnlQuote,
		r.cumulativePnlQuote,
		r.cumulativePnlPercent,
	)

	return r
}

func (r *PrometheusRecorder) ObserveCycle(duration time.Duration, err error) {
	status := "ok"
	if err != nil {
		status = "error"
	}
	r.cyclesTotal.WithLabelValues(status).Inc()
	r.cycleDuration.Observe(duration.Seconds())
}

func (r *PrometheusRecorder) ObserveSignal(signal domain.Signal) {
	r.signalsTotal.WithLabelValues(string(signal)).Inc()
}

func (r *PrometheusRecorder) SetLastPrice(price float64) {
	r.lastPrice.Set(price)
}

func (r *PrometheusRecorder) SetPosition(entryPrice, _ float64, currentPrice float64) {
	r.positionOpen.Set(1)
	if entryPrice <= 0 {
		r.unrealizedPnlPct.Set(0)
		return
	}
	pnlPct := ((currentPrice / entryPrice) - 1) * 100
	r.unrealizedPnlPct.Set(pnlPct)
}

func (r *PrometheusRecorder) ClearPosition() {
	r.positionOpen.Set(0)
	r.unrealizedPnlPct.Set(0)
}

func (r *PrometheusRecorder) ObserveTrade(side domain.OrderSide, mode string, pnlPct, pnlQuote float64) {
	r.tradesTotal.WithLabelValues(string(side), mode).Inc()
	if side == domain.OrderSideSell {
		r.lastTradePnlPct.Set(pnlPct)
		r.lastTradePnlQuote.Set(pnlQuote)
		r.cumulativePnlQuote.Add(pnlQuote)
		r.cumulativePnlPercent.Add(pnlPct)
	}
}
