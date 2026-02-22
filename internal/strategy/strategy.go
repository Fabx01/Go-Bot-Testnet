package strategy

import "github.com/seuprojeto/tradergobinance/internal/domain"

type Strategy interface {
	Name() string
	Evaluate(candles []domain.Candle, position *domain.Position) (domain.Signal, string)
}
