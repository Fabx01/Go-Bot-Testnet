package engine

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/seuprojeto/tradergobinance/internal/api"
	"github.com/seuprojeto/tradergobinance/internal/config"
	"github.com/seuprojeto/tradergobinance/internal/domain"
	logpkg "github.com/seuprojeto/tradergobinance/internal/log"
	"github.com/seuprojeto/tradergobinance/internal/strategy"
)

const apiCallTimeout = 10 * time.Second

type TraderEngine struct {
	cfg      *config.Config
	exchange api.ExchangeClient
	strategy strategy.Strategy
	position *domain.Position
}

func New(cfg *config.Config, exchange api.ExchangeClient, strategy strategy.Strategy) *TraderEngine {
	return &TraderEngine{
		cfg:      cfg,
		exchange: exchange,
		strategy: strategy,
	}
}

func (e *TraderEngine) Run(ctx context.Context) error {
	if err := e.step(ctx); err != nil {
		logpkg.ErrorLogger.Printf("falha no ciclo inicial: %v", err)
	}

	ticker := time.NewTicker(e.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := e.step(ctx); err != nil {
				logpkg.ErrorLogger.Printf("falha no ciclo: %v", err)
			}
		}
	}
}

func (e *TraderEngine) step(ctx context.Context) error {
	apiCtx, cancel := context.WithTimeout(ctx, apiCallTimeout)
	defer cancel()

	candles, err := e.exchange.GetCandles(apiCtx, e.cfg.Symbol, e.cfg.Interval, e.cfg.CandleLimit)
	if err != nil {
		return fmt.Errorf("erro ao buscar candles: %w", sanitizeErr(err))
	}

	currentPrice := candles[len(candles)-1].Close

	signal, reason := e.strategy.Evaluate(candles, e.position)
	logpkg.InfoLogger.Printf("symbol=%s price=%.6f signal=%s reason=%s", e.cfg.Symbol, currentPrice, signal, reason)

	if e.position != nil {
		if currentPrice <= stopLossPrice(e.position.EntryPrice, e.cfg.StopLossPercent) {
			logpkg.InfoLogger.Printf("stop-loss acionado em %.6f", currentPrice)
			signal = domain.SignalSell
		}
		if currentPrice >= takeProfitPrice(e.position.EntryPrice, e.cfg.TakeProfitPercent) {
			logpkg.InfoLogger.Printf("take-profit acionado em %.6f", currentPrice)
			signal = domain.SignalSell
		}
	}

	switch signal {
	case domain.SignalBuy:
		if e.position != nil {
			return nil
		}
		return e.openPosition(ctx, currentPrice)
	case domain.SignalSell:
		if e.position == nil {
			return nil
		}
		return e.closePosition(ctx, currentPrice)
	default:
		return nil
	}
}

func (e *TraderEngine) openPosition(ctx context.Context, price float64) error {
	baseQty := e.cfg.QuoteOrderAmount / price
	if baseQty <= 0 || math.IsNaN(baseQty) || math.IsInf(baseQty, 0) {
		return fmt.Errorf("quantidade base invalida calculada")
	}

	if e.cfg.DryRun {
		e.position = &domain.Position{
			EntryPrice: price,
			BaseQty:    baseQty,
			OpenedAt:   time.Now(),
		}
		logpkg.InfoLogger.Printf("[DRY-RUN] BUY symbol=%s qty=%.8f entry=%.6f", e.cfg.Symbol, baseQty, price)
		return nil
	}

	apiCtx, cancel := context.WithTimeout(ctx, apiCallTimeout)
	defer cancel()
	order, err := e.exchange.PlaceMarketOrder(apiCtx, e.cfg.Symbol, domain.OrderSideBuy, baseQty)
	if err != nil {
		return fmt.Errorf("erro ao enviar ordem BUY: %w", sanitizeErr(err))
	}
	if order.ExecutedQty <= 0 {
		return fmt.Errorf("ordem BUY retornou quantidade executada zero")
	}

	entry := price
	if order.Price > 0 {
		entry = order.Price
	}
	e.position = &domain.Position{
		EntryPrice: entry,
		BaseQty:    order.ExecutedQty,
		OpenedAt:   time.Now(),
	}
	logpkg.InfoLogger.Printf("BUY executado orderId=%d qty=%.8f entry=%.6f", order.OrderID, order.ExecutedQty, entry)
	return nil
}

func (e *TraderEngine) closePosition(ctx context.Context, price float64) error {
	position := e.position
	if position == nil {
		return nil
	}

	if e.cfg.DryRun {
		pnlPct := ((price / position.EntryPrice) - 1) * 100
		pnlQuote := (price - position.EntryPrice) * position.BaseQty
		logpkg.InfoLogger.Printf("[DRY-RUN] SELL symbol=%s qty=%.8f exit=%.6f pnl=%.4f%% pnl_quote=%.6f", e.cfg.Symbol, position.BaseQty, price, pnlPct, pnlQuote)
		e.position = nil
		return nil
	}

	qtyToSell := position.BaseQty * 0.999
	if qtyToSell <= 0 {
		return fmt.Errorf("quantidade para venda invalida")
	}

	apiCtx, cancel := context.WithTimeout(ctx, apiCallTimeout)
	defer cancel()
	order, err := e.exchange.PlaceMarketOrder(apiCtx, e.cfg.Symbol, domain.OrderSideSell, qtyToSell)
	if err != nil {
		return fmt.Errorf("erro ao enviar ordem SELL: %w", sanitizeErr(err))
	}

	exit := price
	if order.Price > 0 {
		exit = order.Price
	}
	pnlPct := ((exit / position.EntryPrice) - 1) * 100
	pnlQuote := (exit - position.EntryPrice) * qtyToSell
	logpkg.InfoLogger.Printf("SELL executado orderId=%d qty=%.8f exit=%.6f pnl=%.4f%% pnl_quote=%.6f", order.OrderID, qtyToSell, exit, pnlPct, pnlQuote)
	e.position = nil
	return nil
}

func stopLossPrice(entryPrice, stopLossPercent float64) float64 {
	return entryPrice * (1 - (stopLossPercent / 100))
}

func takeProfitPrice(entryPrice, takeProfitPercent float64) float64 {
	return entryPrice * (1 + (takeProfitPercent / 100))
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if idx := len(message); idx > 250 {
		message = message[:250] + "..."
	}
	return fmt.Errorf("%s", message)
}
