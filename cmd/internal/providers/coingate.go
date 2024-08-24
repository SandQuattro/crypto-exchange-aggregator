package coingate

import (
	"crypto-exchange-agg/cmd/internal/currency"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type CoinGate struct {
	Client *http.Client
}

func (c *CoinGate) GetAllCurrencies() (string, error) {
	url := fmt.Sprintf("/api/v2/currencies")
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
		log.Fatal(err)
	}

	req.URL.Scheme = "https"
	req.URL.Host = "api.coingate.com"
	req.Header.Add("accept", "text/plain")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}
