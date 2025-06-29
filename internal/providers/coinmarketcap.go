package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/currency"
)

const (
	httpsScheme = "https"
)

type CoinMarketCap struct {
	Logger zerolog.Logger
	Client *http.Client
	Config *config.Config
}

// CoinMarketCapQuote represents a quote from CoinMarketCap API.
type CoinMarketCapQuote struct {
	Price       float64 `json:"price"`
	Volume24h   float64 `json:"volume_24h"`
	MarketCap   float64 `json:"market_cap"`
	LastUpdated string  `json:"last_updated"`
}

// CoinMarketCapData represents cryptocurrency data.
type CoinMarketCapData struct {
	ID     int                           `json:"id"`
	Name   string                        `json:"name"`
	Symbol string                        `json:"symbol"`
	Quote  map[string]CoinMarketCapQuote `json:"quote"`
}

// CoinMarketCapResponse represents the API response.
type CoinMarketCapResponse struct {
	Status struct {
		Timestamp    string `json:"timestamp"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
	Data map[string]CoinMarketCapData `json:"data"`
}

// GetLatestQuotes получает последние котировки для указанных символов.
func (c *CoinMarketCap) GetLatestQuotes(
	ctx context.Context,
	symbols []currency.Cryptocurrency,
	convert string,
) (*CoinMarketCapResponse, error) {
	symbolsStr := make([]string, len(symbols))
	for i, symbol := range symbols {
		symbolsStr[i] = symbol.String()
	}

	url := fmt.Sprintf("/v1/cryptocurrency/quotes/latest?symbol=%s&convert=%s",
		strings.Join(symbolsStr, ","), convert)

	body, err := c.callRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response CoinMarketCapResponse
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", unmarshalErr)
	}

	if response.Status.ErrorCode != 0 {
		return nil, fmt.Errorf("api error %d: %s", response.Status.ErrorCode, response.Status.ErrorMessage)
	}

	return &response, nil
}

// GetQuote получает котировку для одной валюты.
func (c *CoinMarketCap) GetQuote(ctx context.Context, symbol currency.Cryptocurrency, convert string) (float64, error) {
	response, err := c.GetLatestQuotes(ctx, []currency.Cryptocurrency{symbol}, convert)
	if err != nil {
		return 0, err
	}

	symbolStr := symbol.String()
	data, exists := response.Data[symbolStr]
	if !exists {
		return 0, fmt.Errorf("symbol %s not found in response", symbolStr)
	}

	quote, exists := data.Quote[convert]
	if !exists {
		return 0, fmt.Errorf("convert currency %s not found for %s", convert, symbolStr)
	}

	return quote.Price, nil
}

// GetQuoteString получает котировку как строку (для совместимости с другими провайдерами).
func (c *CoinMarketCap) GetQuoteString(
	ctx context.Context,
	symbol currency.Cryptocurrency,
	convert string,
) (string, error) {
	price, err := c.GetQuote(ctx, symbol, convert)
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(price, 'f', 8, 64), nil
}

// GetRate реализует интерфейс QuoteProvider.
func (c *CoinMarketCap) GetRate(ctx context.Context, from, to currency.Cryptocurrency) (string, error) {
	return c.GetQuoteString(ctx, from, to.String())
}

// GetRates получает курсы для нескольких валют (реализация RateProvider).
func (c *CoinMarketCap) GetRates(
	ctx context.Context,
	symbols []currency.Cryptocurrency,
	convert string,
) (map[string]string, error) {
	response, err := c.GetLatestQuotes(ctx, symbols, convert)
	if err != nil {
		return nil, err
	}

	rates := make(map[string]string)
	for _, symbol := range symbols {
		symbolStr := symbol.String()
		data, exists := response.Data[symbolStr]
		if !exists {
			c.Logger.Warn().
				Str("symbol", symbolStr).
				Msg("Symbol not found in CoinMarketCap response")
			continue
		}

		quote, exists := data.Quote[convert]
		if !exists {
			c.Logger.Warn().
				Str("symbol", symbolStr).
				Str("convert", convert).
				Msg("Convert currency not found in quote")
			continue
		}

		rates[symbolStr] = strconv.FormatFloat(quote.Price, 'f', 8, 64)
	}

	return rates, nil
}

// callRequest выполняет HTTP запрос к CoinMarketCap API.
func (c *CoinMarketCap) callRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.Scheme = httpsScheme
	req.URL.Host = "pro-api.coinmarketcap.com"
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Cmc_pro_api_key", c.Config.Key)

	c.Logger.Debug().Str("url", req.URL.String()).Msg("Making CoinMarketCap API request")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			c.Logger.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		c.Logger.Error().
			Int("status_code", res.StatusCode).
			Str("response", string(body)).
			Msg("CoinMarketCap API returned error")
		return nil, fmt.Errorf("api response: %d, body: %s", res.StatusCode, string(body))
	}

	return body, nil
}
