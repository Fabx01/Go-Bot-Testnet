package config

import "testing"

func TestLoadDefaultsToSafeModes(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "key")
	t.Setenv("BINANCE_API_SECRET", "secret")
	t.Setenv("BINANCE_TESTNET", "")
	t.Setenv("BOT_DRY_RUN", "")
	t.Setenv("BINANCE_ALLOW_MAINNET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.UseTestnet {
		t.Fatalf("expected UseTestnet=true by default")
	}
	if !cfg.DryRun {
		t.Fatalf("expected DryRun=true by default")
	}
	if !cfg.Observability {
		t.Fatalf("expected Observability=true by default")
	}
	if cfg.ObservabilityAddr == "" {
		t.Fatalf("expected ObservabilityAddr default value")
	}
}

func TestLoadBlocksMainnetWithoutExplicitAllow(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "key")
	t.Setenv("BINANCE_API_SECRET", "secret")
	t.Setenv("BINANCE_TESTNET", "false")
	t.Setenv("BINANCE_ALLOW_MAINNET", "false")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error when mainnet is not explicitly allowed")
	}
}
