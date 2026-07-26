package env_test

import (
	"os"
	"testing"

	"gitlab.com/massimo-ua/projecta/pkg/env"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns env value when variable is set", func(t *testing.T) {
		const key = "TEST_GET_ENV_VAR"
		const expected = "custom_value"

		_ = os.Setenv(key, expected)
		defer os.Unsetenv(key)

		got := env.GetEnv(key, "default_value")
		if got != expected {
			t.Errorf("GetEnv(%q) = %q; want %q", key, got, expected)
		}
	})

	t.Run("returns default value when variable is unset", func(t *testing.T) {
		const key = "TEST_GET_ENV_UNSET_VAR"
		const defaultVal = "default_value"

		os.Unsetenv(key)

		got := env.GetEnv(key, defaultVal)
		if got != defaultVal {
			t.Errorf("GetEnv(%q) = %q; want %q", key, got, defaultVal)
		}
	})
}
