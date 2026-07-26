package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	defaultNBUAPIURL = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
	defaultTTL       = 30 * time.Minute
)

type nbuRateItem struct {
	R030         int     `json:"r030"`
	Txt          string  `json:"txt"`
	Rate         float64 `json:"rate"`
	CC           string  `json:"cc"`
	ExchangeDate string  `json:"exchangedate"`
}

type NBUCurrencyRateProviderOptions struct {
	APIURL             string
	SupportedCurrencies []string
	CacheTTL           time.Duration
	HTTPClient         *http.Client
}

type NBUCurrencyRateProvider struct {
	apiURL              string
	supportedCurrencies map[string]bool
	cacheTTL            time.Duration
	httpClient          *http.Client

	mu          sync.RWMutex
	rates       map[string]float64
	lastFetched time.Time
}

func NewNBUCurrencyRateProvider(opts NBUCurrencyRateProviderOptions) *NBUCurrencyRateProvider {
	apiURL := opts.APIURL
	if apiURL == "" {
		apiURL = defaultNBUAPIURL
	}

	ttl := opts.CacheTTL
	if ttl <= 0 {
		ttl = defaultTTL
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	supported := make(map[string]bool)
	if len(opts.SupportedCurrencies) == 0 {
		opts.SupportedCurrencies = []string{"UAH", "USD", "EUR", "PLN"}
	}
	for _, c := range opts.SupportedCurrencies {
		supported[strings.ToUpper(strings.TrimSpace(c))] = true
	}
	// Always support UAH as base
	supported["UAH"] = true

	return &NBUCurrencyRateProvider{
		apiURL:              apiURL,
		supportedCurrencies: supported,
		cacheTTL:            ttl,
		httpClient:          client,
		rates:               make(map[string]float64),
	}
}

func (p *NBUCurrencyRateProvider) Convert(currencyA Currency, currencyB Currency) (Currency, error) {
	codeA := strings.ToUpper(strings.TrimSpace(currencyA.Code))
	codeB := strings.ToUpper(strings.TrimSpace(currencyB.Code))

	if !p.supportedCurrencies[codeA] {
		return Currency{}, exceptions.NewValidationException(fmt.Sprintf("unsupported source currency: %s", codeA), nil)
	}
	if !p.supportedCurrencies[codeB] {
		return Currency{}, exceptions.NewValidationException(fmt.Sprintf("unsupported target currency: %s", codeB), nil)
	}

	if codeA == codeB {
		return Currency{Amount: currencyA.Amount, Code: codeB}, nil
	}

	rates, err := p.getRates()
	if err != nil {
		return Currency{}, err
	}

	rateA, okA := rates[codeA]
	rateB, okB := rates[codeB]

	if !okA {
		return Currency{}, exceptions.NewInternalException(fmt.Sprintf("rate for currency %s not found", codeA), nil)
	}
	if !okB {
		return Currency{}, exceptions.NewInternalException(fmt.Sprintf("rate for currency %s not found", codeB), nil)
	}

	if rateB <= 0 {
		return Currency{}, exceptions.NewInternalException(fmt.Sprintf("invalid rate for currency %s", codeB), nil)
	}

	// Formula: amount * (rateA / rateB)
	converted := float64(currencyA.Amount) * (rateA / rateB)
	resultAmount := int64(math.Round(converted))

	return Currency{
		Amount: resultAmount,
		Code:   codeB,
	}, nil
}

func (p *NBUCurrencyRateProvider) getRates() (map[string]float64, error) {
	p.mu.RLock()
	if len(p.rates) > 0 && p.isCacheValid() {
		defer p.mu.RUnlock()
		return p.rates, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check inside write lock
	if len(p.rates) > 0 && p.isCacheValid() {
		return p.rates, nil
	}

	rates, err := p.fetchRatesFromNBU()
	if err != nil {
		// If fetch fails but we have stale rates in cache, fallback to stale rates
		if len(p.rates) > 0 {
			return p.rates, nil
		}
		return nil, err
	}

	p.rates = rates
	p.lastFetched = time.Now()
	return p.rates, nil
}

func (p *NBUCurrencyRateProvider) isCacheValid() bool {
	if p.lastFetched.IsZero() {
		return false
	}

	now := time.Now()
	if now.Sub(p.lastFetched) > p.cacheTTL {
		return false
	}

	// Check NBU 15:30 Europe/Kyiv release rule:
	// If lastFetched was before 15:30 Kyiv time and current time is after 15:30 Kyiv time on the same or later day,
	// invalidate cache to pull newly published rates.
	kyivLoc, err := time.LoadLocation("Europe/Kyiv")
	if err == nil {
		nowKyiv := now.In(kyivLoc)
		fetchedKyiv := p.lastFetched.In(kyivLoc)

		cutoffToday := time.Date(nowKyiv.Year(), nowKyiv.Month(), nowKyiv.Day(), 15, 30, 0, 0, kyivLoc)
		if nowKyiv.After(cutoffToday) && fetchedKyiv.Before(cutoffToday) {
			return false
		}
	}

	return true
}

func (p *NBUCurrencyRateProvider) fetchRatesFromNBU() (map[string]float64, error) {
	resp, err := p.httpClient.Get(p.apiURL)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch NBU exchange rates", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, exceptions.NewInternalException(fmt.Sprintf("NBU API returned status %d: %s", resp.StatusCode, string(bodyBytes)), nil)
	}

	var items []nbuRateItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, exceptions.NewInternalException("failed to decode NBU exchange rates response", err)
	}

	rates := make(map[string]float64)
	rates["UAH"] = 1.0 // UAH is always 1.0

	for _, item := range items {
		code := strings.ToUpper(strings.TrimSpace(item.CC))
		if p.supportedCurrencies[code] {
			rates[code] = item.Rate
		}
	}

	return rates, nil
}
