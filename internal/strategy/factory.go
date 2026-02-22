package strategy

import (
	"fmt"
	"strings"

	"github.com/seuprojeto/tradergobinance/internal/config"
)

func Build(cfg *config.Config) (Strategy, error) {
	switch strings.ToLower(cfg.StrategyName) {
	case "ema_rsi":
		return NewEMARSI(cfg.EMAShortPeriod, cfg.EMALongPeriod, cfg.RSIPeriod, cfg.RSIOverboughtLevel), nil
	default:
		return nil, fmt.Errorf("estrategia nao suportada: %s", cfg.StrategyName)
	}
}
