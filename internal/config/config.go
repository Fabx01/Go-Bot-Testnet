package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BinanceAPIKey      string
	BinanceAPISecret   string
	UseTestnet         bool
	AllowMainnet       bool
	DryRun             bool
	Symbol             string
	Interval           string
	CandleLimit        int
	PollInterval       time.Duration
	QuoteOrderAmount   float64
	StopLossPercent    float64
	TakeProfitPercent  float64
	StrategyName       string
	EMAShortPeriod     int
	EMALongPeriod      int
	RSIPeriod          int
	RSIOverboughtLevel float64
}

func Load() (*Config, error) {
	cfg := &Config{
		BinanceAPIKey:      strings.TrimSpace(os.Getenv("BINANCE_API_KEY")),
		BinanceAPISecret:   strings.TrimSpace(os.Getenv("BINANCE_API_SECRET")),
		UseTestnet:         parseBoolWithDefault("BINANCE_TESTNET", true),
		AllowMainnet:       parseBoolWithDefault("BINANCE_ALLOW_MAINNET", false),
		DryRun:             parseBoolWithDefault("BOT_DRY_RUN", true),
		Symbol:             strings.ToUpper(parseStringWithDefault("BOT_SYMBOL", "BTCUSDT")),
		Interval:           parseStringWithDefault("BOT_INTERVAL", "1m"),
		CandleLimit:        parseIntWithDefault("BOT_CANDLE_LIMIT", 200),
		PollInterval:       parseDurationWithDefault("BOT_POLL_INTERVAL", 20*time.Second),
		QuoteOrderAmount:   parseFloatWithDefault("BOT_QUOTE_ORDER_AMOUNT", 20),
		StopLossPercent:    parseFloatWithDefault("BOT_STOP_LOSS_PCT", 1.5),
		TakeProfitPercent:  parseFloatWithDefault("BOT_TAKE_PROFIT_PCT", 3.0),
		StrategyName:       parseStringWithDefault("BOT_STRATEGY", "ema_rsi"),
		EMAShortPeriod:     parseIntWithDefault("EMA_SHORT_PERIOD", 9),
		EMALongPeriod:      parseIntWithDefault("EMA_LONG_PERIOD", 21),
		RSIPeriod:          parseIntWithDefault("RSI_PERIOD", 14),
		RSIOverboughtLevel: parseFloatWithDefault("RSI_OVERBOUGHT", 70),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.BinanceAPIKey == "" || c.BinanceAPISecret == "" {
		return fmt.Errorf("BINANCE_API_KEY e BINANCE_API_SECRET sao obrigatorias")
	}
	if !c.UseTestnet && !c.AllowMainnet {
		return fmt.Errorf("mainnet bloqueada: defina BINANCE_ALLOW_MAINNET=true para liberar")
	}
	if c.CandleLimit < 50 {
		return fmt.Errorf("BOT_CANDLE_LIMIT deve ser >= 50")
	}
	if c.PollInterval < 5*time.Second {
		return fmt.Errorf("BOT_POLL_INTERVAL deve ser >= 5s")
	}
	if c.QuoteOrderAmount <= 0 {
		return fmt.Errorf("BOT_QUOTE_ORDER_AMOUNT deve ser > 0")
	}
	if c.StopLossPercent <= 0 || c.TakeProfitPercent <= 0 {
		return fmt.Errorf("BOT_STOP_LOSS_PCT e BOT_TAKE_PROFIT_PCT devem ser > 0")
	}
	if c.EMAShortPeriod <= 1 || c.EMALongPeriod <= 1 || c.RSIPeriod <= 1 {
		return fmt.Errorf("periodos de indicadores devem ser > 1")
	}
	if c.EMAShortPeriod >= c.EMALongPeriod {
		return fmt.Errorf("EMA_SHORT_PERIOD deve ser menor que EMA_LONG_PERIOD")
	}
	if c.RSIOverboughtLevel <= 50 || c.RSIOverboughtLevel >= 95 {
		return fmt.Errorf("RSI_OVERBOUGHT deve ficar entre 50 e 95")
	}
	if c.StrategyName == "" {
		return fmt.Errorf("BOT_STRATEGY nao pode ser vazio")
	}
	return nil
}

func parseStringWithDefault(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func parseBoolWithDefault(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseIntWithDefault(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseFloatWithDefault(key string, defaultValue float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
