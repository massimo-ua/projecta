package currency

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNBUCurrencyRateProvider_Convert(t *testing.T) {
	mockResponse := `[
		{"r030":840,"txt":"Долар США","rate":40.00,"cc":"USD","exchangedate":"26.07.2026"},
		{"r030":978,"txt":"Євро","rate":44.00,"cc":"EUR","exchangedate":"26.07.2026"},
		{"r030":985,"txt":"Злотий","rate":10.00,"cc":"PLN","exchangedate":"26.07.2026"}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mockResponse)
	}))
	defer server.Close()

	provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
		APIURL:              server.URL,
		SupportedCurrencies: []string{"UAH", "USD", "EUR", "PLN"},
		CacheTTL:            1 * time.Minute,
	})

	tests := []struct {
		name      string
		from      Currency
		to        Currency
		expected  int64
		expectErr bool
	}{
		{
			name:      "Same currency (UAH -> UAH)",
			from:      Currency{Amount: 10000, Code: "UAH"},
			to:        Currency{Code: "UAH"},
			expected:  10000,
			expectErr: false,
		},
		{
			name:      "USD to UAH (100 USD @ 40.00)",
			from:      Currency{Amount: 10000, Code: "USD"},
			to:        Currency{Code: "UAH"},
			expected:  400000,
			expectErr: false,
		},
		{
			name:      "UAH to USD (400 UAH @ 40.00)",
			from:      Currency{Amount: 40000, Code: "UAH"},
			to:        Currency{Code: "USD"},
			expected:  1000,
			expectErr: false,
		},
		{
			name:      "USD to EUR (110 USD -> EUR, 40/44)",
			from:      Currency{Amount: 11000, Code: "USD"},
			to:        Currency{Code: "EUR"},
			expected:  10000,
			expectErr: false,
		},
		{
			name:      "PLN to UAH (50 PLN @ 10.00)",
			from:      Currency{Amount: 5000, Code: "PLN"},
			to:        Currency{Code: "UAH"},
			expected:  50000,
			expectErr: false,
		},
		{
			name:      "Unsupported source currency (GBP)",
			from:      Currency{Amount: 1000, Code: "GBP"},
			to:        Currency{Code: "UAH"},
			expectErr: true,
		},
		{
			name:      "Unsupported target currency (GBP)",
			from:      Currency{Amount: 1000, Code: "USD"},
			to:        Currency{Code: "GBP"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Convert(tt.from, tt.to)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Amount != tt.expected {
				t.Errorf("expected amount %d, got %d", tt.expected, result.Amount)
			}
			if result.Code != tt.to.Code {
				t.Errorf("expected code %s, got %s", tt.to.Code, result.Code)
			}
		})
	}
}

func TestNBUCurrencyRateProvider_Defaults(t *testing.T) {
	provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{})
	if provider.apiURL == "" {
		t.Errorf("expected default API URL, got empty")
	}
	if provider.cacheTTL <= 0 {
		t.Errorf("expected positive cache TTL, got %v", provider.cacheTTL)
	}
}

func TestNBUCurrencyRateProvider_ErrorHandling(t *testing.T) {
	t.Run("HTTP 500 Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "internal server error")
		}))
		defer server.Close()

		provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
			APIURL: server.URL,
		})

		_, err := provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "UAH"})
		if err == nil {
			t.Fatalf("expected error on HTTP 500, got none")
		}
	})

	t.Run("Invalid JSON Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, "not valid json")
		}))
		defer server.Close()

		provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
			APIURL: server.URL,
		})

		_, err := provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "UAH"})
		if err == nil {
			t.Fatalf("expected error on invalid JSON, got none")
		}
	})

	t.Run("Fallback to stale rates when API fails", func(t *testing.T) {
		status := http.StatusOK
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if status != http.StatusOK {
				w.WriteHeader(status)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `[{"r030":840,"txt":"Долар США","rate":40.00,"cc":"USD","exchangedate":"26.07.2026"}]`)
		}))
		defer server.Close()

		provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
			APIURL:   server.URL,
			CacheTTL: 1 * time.Millisecond,
		})

		// First call succeeds and caches
		_, err := provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "UAH"})
		if err != nil {
			t.Fatalf("unexpected error on initial fetch: %v", err)
		}

		time.Sleep(5 * time.Millisecond) // expire cache

		// Subsequent call fails at server level, but uses stale cache
		status = http.StatusInternalServerError
		res, err := provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "UAH"})
		if err != nil {
			t.Fatalf("expected fallback to stale rates, got error: %v", err)
		}
		if res.Amount != 4000 {
			t.Errorf("expected 4000, got %d", res.Amount)
		}
	})

	t.Run("Missing rate for supported currency", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `[{"r030":840,"txt":"USD","rate":40.00,"cc":"USD","exchangedate":"26.07.2026"}]`)
		}))
		defer server.Close()

		provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
			APIURL:              server.URL,
			SupportedCurrencies: []string{"USD", "EUR"},
		})

		_, err := provider.Convert(Currency{Amount: 100, Code: "EUR"}, Currency{Code: "USD"})
		if err == nil {
			t.Fatalf("expected error when source rate is missing")
		}

		_, err = provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "EUR"})
		if err == nil {
			t.Fatalf("expected error when target rate is missing")
		}
	})

	t.Run("Invalid zero rate for target currency", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `[
				{"r030":840,"txt":"USD","rate":40.00,"cc":"USD","exchangedate":"26.07.2026"},
				{"r030":978,"txt":"EUR","rate":0.00,"cc":"EUR","exchangedate":"26.07.2026"}
			]`)
		}))
		defer server.Close()

		provider := NewNBUCurrencyRateProvider(NBUCurrencyRateProviderOptions{
			APIURL:              server.URL,
			SupportedCurrencies: []string{"USD", "EUR"},
		})

		_, err := provider.Convert(Currency{Amount: 100, Code: "USD"}, Currency{Code: "EUR"})
		if err == nil {
			t.Fatalf("expected error when rate <= 0")
		}
	})
}
