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
			expected:  400000, // 10000 * 40.0 = 400000
			expectErr: false,
		},
		{
			name:      "UAH to USD (400 UAH @ 40.00)",
			from:      Currency{Amount: 40000, Code: "UAH"},
			to:        Currency{Code: "USD"},
			expected:  1000, // 40000 / 40.0 = 1000
			expectErr: false,
		},
		{
			name:      "USD to EUR (110 USD -> EUR, 40/44)",
			from:      Currency{Amount: 11000, Code: "USD"},
			to:        Currency{Code: "EUR"},
			expected:  10000, // 11000 * (40 / 44) = 10000
			expectErr: false,
		},
		{
			name:      "PLN to UAH (50 PLN @ 10.00)",
			from:      Currency{Amount: 5000, Code: "PLN"},
			to:        Currency{Code: "UAH"},
			expected:  50000, // 5000 * 10 = 50000
			expectErr: false,
		},
		{
			name:      "Unsupported currency (GBP)",
			from:      Currency{Amount: 1000, Code: "GBP"},
			to:        Currency{Code: "UAH"},
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
