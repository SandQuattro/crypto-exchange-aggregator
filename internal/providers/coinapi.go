package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/currency"
)

type CoinApi struct {
	Client *http.Client
	Config *config.Config
}

func (c *CoinApi) GetAllCurrencies() (string, error) {
	url := "/v1/assets"
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinApi) GetUSDRates(from []currency.Cryptocurrency) (string, error) {
	var assets strings.Builder
	for _, asset := range from {
		assets.WriteString(asset.String())
		assets.WriteString(";")
	}

	url := fmt.Sprintf("/v1/assets?filter_asset_id=%s", assets.String())
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinApi) GetRate(cryptocurrency currency.Cryptocurrency) (string, error) {
	url := fmt.Sprintf("/v1/exchangerate/%s", cryptocurrency.String())
	request, err := c.callRequest(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(request)), nil
}

func (c *CoinApi) callRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.URL.Scheme = "https"
	req.URL.Host = "rest.coinapi.io"
	req.Header.Add("accept", "text/plain")
	req.Header.Add("Authorization", c.Config.CoinAPI.Key)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	bodyMap := make(map[any]any)
	err = json.Unmarshal(body, &bodyMap)

	var e *json.UnmarshalTypeError
	if errors.As(err, &e) {
		if e.Value == "array" {
			return body, nil
		}
	}

	if err != nil {
		return nil, err
	}

	if val, ok := bodyMap["error"]; ok {
		return nil, fmt.Errorf("api response: %d, error:%s", res.StatusCode, val)
	}

	return body, err
}
