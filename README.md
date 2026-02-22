# TraderGoBinance

Bot de trading em Go para Binance com foco em execucao continua (VPS), codigo desacoplado e modo seguro por padrao.

## Status atual

- Exchange desacoplada por interface (`internal/api`)
- Engine com loop continuo e desligamento limpo por sinal (`internal/engine`)
- Estrategia plugavel (`internal/strategy`)
- Gerenciamento de risco minimo: `stop-loss` e `take-profit`
- `testnet` e `dry-run` ativados por padrao para reduzir risco operacional

## Arquitetura

- `cmd/main.go`: bootstrap e wiring
- `internal/config`: leitura/validacao das variaveis de ambiente
- `internal/api`: adaptador Binance (`go-binance`)
- `internal/domain`: tipos de dominio (candles, sinal, posicao, ordem)
- `internal/strategy`: estrategias e indicadores
- `internal/engine`: ciclo de decisao e execucao

## Estrategia implementada

- `ema_rsi`
- Compra: cruzamento de EMA curta acima da EMA longa, com RSI abaixo do limite de sobrecompra.
- Venda: cruzamento de EMA curta para baixo ou RSI acima do limite de sobrecompra.
- Protecoes extras: fechamento por `stop-loss` e `take-profit`.

## Configuracao

Copie `.env.example` para `.env` e ajuste os valores.

Variaveis principais:

- `BINANCE_API_KEY`
- `BINANCE_API_SECRET`
- `BINANCE_TESTNET` (`true/false`, default `true`)
- `BINANCE_ALLOW_MAINNET` (default `false`)
- `BOT_DRY_RUN` (`true/false`, default `true`)
- `BOT_SYMBOL` (default `BTCUSDT`)
- `BOT_INTERVAL` (default `1m`)
- `BOT_POLL_INTERVAL` (default `20s`)
- `BOT_QUOTE_ORDER_AMOUNT` (default `20`)
- `BOT_STOP_LOSS_PCT` (default `1.5`)
- `BOT_TAKE_PROFIT_PCT` (default `3.0`)
- `BOT_STRATEGY` (default `ema_rsi`)
- `EMA_SHORT_PERIOD` (default `9`)
- `EMA_LONG_PERIOD` (default `21`)
- `RSI_PERIOD` (default `14`)
- `RSI_OVERBOUGHT` (default `70`)

## Rodando local

```sh
go mod tidy
go build -o traderbot ./cmd
./traderbot
```

## Deploy em VPS (DigitalOcean)

1. Crie uma Droplet Ubuntu 22.04+.
2. Instale Go 1.21+ e copie o projeto.
3. Configure `.env` com `testnet=true` e `dry_run=true` no inicio.
4. Compile:

```sh
go build -o /opt/traderbot/traderbot ./cmd
```

5. Crie unit `systemd` (`/etc/systemd/system/traderbot.service`):

```ini
[Unit]
Description=TraderGoBinance Bot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/traderbot
EnvironmentFile=/opt/traderbot/.env
ExecStart=/opt/traderbot/traderbot
Restart=always
RestartSec=5
User=trader

[Install]
WantedBy=multi-user.target
```

6. Habilite o servico:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now traderbot
sudo journalctl -u traderbot -f
```

## Observacoes de risco

- Nao existe estrategia "melhor" universal ou garantia de lucro.
- Antes de usar dinheiro real: rode por dias em `testnet` e depois em `dry-run` na mainnet apenas para observar sinais.
- Recomendado incluir backtest e limite diario de perda antes de operar em live.
