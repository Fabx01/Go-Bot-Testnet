package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/seuprojeto/tradergobinance/internal/domain"
)

const binanceTestnetURL = "https://testnet.binance.vision"

type ExchangeClient interface {
	GetAccountBalances(ctx context.Context) ([]domain.AccountBalance, error)
	GetCandles(ctx context.Context, symbol, interval string, limit int) ([]domain.Candle, error)
	GetPrice(ctx context.Context, symbol string) (float64, error)
	PlaceMarketOrder(ctx context.Context, symbol string, side domain.OrderSide, quantity float64) (*domain.OrderResult, error)
}

type BinanceClient struct {
	client *binance.Client
}

func NewBinanceClient(apiKey, apiSecret string, testnet bool) *BinanceClient {
	client := binance.NewClient(apiKey, apiSecret)
	if testnet {
		client.BaseURL = binanceTestnetURL
	}
	return &BinanceClient{client: client}
}

func (b *BinanceClient) GetAccountBalances(ctx context.Context) ([]domain.AccountBalance, error) {
	account, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}
	balances := make([]domain.AccountBalance, 0, len(account.Balances))
	for _, bal := range account.Balances {
		free, parseErr := strconv.ParseFloat(bal.Free, 64)
		if parseErr != nil {
			continue
		}
		if free <= 0 {
			continue
		}
		balances = append(balances, domain.AccountBalance{Asset: bal.Asset, Free: free})
	}
	return balances, nil
}

func (b *BinanceClient) GetCandles(ctx context.Context, symbol, interval string, limit int) ([]domain.Candle, error) {
	klines, err := b.client.NewKlinesService().
		Symbol(strings.ToUpper(symbol)).
		Interval(interval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	candles := make([]domain.Candle, 0, len(klines))
	for _, k := range klines {
		open, errOpen := strconv.ParseFloat(k.Open, 64)
		high, errHigh := strconv.ParseFloat(k.High, 64)
		low, errLow := strconv.ParseFloat(k.Low, 64)
		closePrice, errClose := strconv.ParseFloat(k.Close, 64)
		volume, errVolume := strconv.ParseFloat(k.Volume, 64)
		if errOpen != nil || errHigh != nil || errLow != nil || errClose != nil || errVolume != nil {
			continue
		}
		candles = append(candles, domain.Candle{
			OpenTime:  time.UnixMilli(k.OpenTime),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: time.UnixMilli(k.CloseTime),
		})
	}
	if len(candles) == 0 {
		return nil, fmt.Errorf("nenhum candle valido retornado")
	}
	return candles, nil
}

func (b *BinanceClient) GetPrice(ctx context.Context, symbol string) (float64, error) {
	prices, err := b.client.NewListPricesService().Symbol(strings.ToUpper(symbol)).Do(ctx)
	if err != nil {
		return 0, err
	}
	if len(prices) == 0 {
		return 0, fmt.Errorf("preco nao encontrado para %s", symbol)
	}
	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func (b *BinanceClient) PlaceMarketOrder(ctx context.Context, symbol string, side domain.OrderSide, quantity float64) (*domain.OrderResult, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity invalida: %.8f", quantity)
	}

	order, err := b.client.NewCreateOrderService().
		Symbol(strings.ToUpper(symbol)).
		Side(binance.SideType(side)).
		Type(binance.OrderTypeMarket).
		Quantity(formatQuantity(quantity)).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	executedQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	price := 0.0
	if executedQty > 0 {
		cummulativeQuoteQty, parseErr := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)
		if parseErr == nil && cummulativeQuoteQty > 0 {
			price = cummulativeQuoteQty / executedQty
		}
	}

	return &domain.OrderResult{
		Symbol:      order.Symbol,
		Side:        side,
		ExecutedQty: executedQty,
		Price:       price,
		OrderID:     order.OrderID,
	}, nil
}

func formatQuantity(quantity float64) string {
	formatted := strconv.FormatFloat(quantity, 'f', 8, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	if formatted == "" {
		return "0"
	}
	return formatted
}
