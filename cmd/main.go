package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/seuprojeto/tradergobinance/internal/api"
	"github.com/seuprojeto/tradergobinance/internal/config"
	"github.com/seuprojeto/tradergobinance/internal/engine"
	logpkg "github.com/seuprojeto/tradergobinance/internal/log"
	"github.com/seuprojeto/tradergobinance/internal/observability"
	"github.com/seuprojeto/tradergobinance/internal/strategy"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logpkg.ErrorLogger.Fatalf("falha ao carregar config: %v", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var recorder observability.Recorder = observability.NewNoopRecorder()
	if cfg.Observability {
		obsServer := observability.NewServer(cfg.ObservabilityAddr)
		recorder = obsServer.Recorder()
		logpkg.InfoLogger.Printf("observability ativa em %s (/metrics, /healthz)", obsServer.Addr())
		go func() {
			if runErr := obsServer.Run(ctx); runErr != nil {
				logpkg.ErrorLogger.Printf("falha na observability: %v", runErr)
				cancel()
			}
		}()
	}

	exchange := api.NewBinanceClient(cfg.BinanceAPIKey, cfg.BinanceAPISecret, cfg.UseTestnet)
	strat, err := strategy.Build(cfg)
	if err != nil {
		logpkg.ErrorLogger.Fatalf("falha ao montar estrategia: %v", err)
	}

	trader := engine.New(cfg, exchange, strat, recorder)
	logpkg.InfoLogger.Printf("bot iniciado symbol=%s interval=%s testnet=%t dry_run=%t strategy=%s", cfg.Symbol, cfg.Interval, cfg.UseTestnet, cfg.DryRun, strat.Name())

	if err := trader.Run(ctx); err != nil {
		logpkg.ErrorLogger.Fatalf("erro fatal no engine: %v", err)
	}

	logpkg.InfoLogger.Println("bot finalizado")
}
