package providers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"crypto-exchange-agg/internal/currency"

	"github.com/rs/zerolog"
)

type CoinGate struct {
	Logger zerolog.Logger
	Client *http.Client
}

func (c *CoinGate) GetAllCurrencies() (string, error) {
	url := "/api/v2/currencies"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetAllRates() (string, error) {
	url := "/v2/rates"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetAllMerchantRates() (string, error) {
	url := "/v2/rates/merchant"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetAllTraderRates() (string, error) {
	url := "/v2/rates/trader"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetTraderBuyRates() (string, error) {
	url := "v2/rates/trader/buy"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetTraderSellRates() (string, error) {
	url := "v2/rates/trader/sell"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetMerchantRate(from, to currency.Cryptocurrency) (string, error) {
	url := fmt.Sprintf("/api/v2/rates/merchant/%s/%s", from, to)
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetTraderBuy(from, to currency.Cryptocurrency) (string, error) {
	url := fmt.Sprintf("/api/v2/rates/trader/buy/%s/%s", from, to)
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) GetTraderSell(from, to currency.Cryptocurrency) (string, error) {
	url := fmt.Sprintf("/api/v2/rates/trader/sell/%s/%s", from, to)
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinGate) callRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("Failed to create HTTP request")
		return nil, err
	}

	req.URL.Scheme = "https"
	req.URL.Host = "api.coingate.com"
	req.Header.Add("accept", "text/plain")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			c.Logger.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("api response: %d, error:%s", res.StatusCode, string(body))
	}

	return body, err
}
