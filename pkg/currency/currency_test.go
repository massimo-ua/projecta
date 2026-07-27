package currency

import "testing"

func TestCurrency(t *testing.T) {
	c := NewCurrency(100, "USD")
	if c.Amount != 100 {
		t.Errorf("expected 100, got %d", c.Amount)
	}
	if c.Code != "USD" {
		t.Errorf("expected USD, got %s", c.Code)
	}
	if c.String() != "100 USD" {
		t.Errorf("expected '100 USD', got '%s'", c.String())
	}
}
