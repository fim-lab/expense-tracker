package stockprice

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

// HTTPPriceFetcher fetches a stock's current price by substituting $ticker
// into a configurable URL template and parsing the response.
type HTTPPriceFetcher struct {
	urlTemplate string
	client      *http.Client
}

func NewHTTPPriceFetcher(urlTemplate string) *HTTPPriceFetcher {
	log.Printf("Setup with url %v", urlTemplate)
	return &HTTPPriceFetcher{
		urlTemplate: urlTemplate,
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

// chartResponse mirrors the subset of the response shape this fetcher
// currently expects: {"chart":{"result":[{"meta":{"regularMarketPrice":...}}],"error":null}}.
type chartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
		} `json:"result"`
		Error *struct {
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func (f *HTTPPriceFetcher) FetchPrice(ticker string) (int, error) {
	if f.urlTemplate == "" {
		return 0, domain.ErrPriceFetchFailed
	}

	reqURL := strings.ReplaceAll(f.urlTemplate, "$ticker", url.QueryEscape(ticker))
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrPriceFetchFailed, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; expense-tracker/1.0)")

	resp, err := f.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrPriceFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: upstream status %d", domain.ErrPriceFetchFailed, resp.StatusCode)
	}

	var parsed chartResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("%w: %v", domain.ErrPriceFetchFailed, err)
	}
	if parsed.Chart.Error != nil || len(parsed.Chart.Result) == 0 {
		return 0, domain.ErrPriceFetchFailed
	}

	price := parsed.Chart.Result[0].Meta.RegularMarketPrice
	if price <= 0 {
		return 0, domain.ErrPriceFetchFailed
	}

	return int(math.Round(price * 100)), nil
}
