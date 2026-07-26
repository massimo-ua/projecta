package logger_test

import (
	"errors"
	"testing"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/pkg/logger"
)

func TestLogger(t *testing.T) {
	t.Run("Info logs with fields and nil fields", func(t *testing.T) {
		log := logger.New("password", "token")
		log.Info("test info with nil fields", nil)
		log.Info("test info with fields", map[string]any{
			"user_id":  "123",
			"password": "secret_password",
		})
	})

	t.Run("Error logs with standard error and exception", func(t *testing.T) {
		log := logger.New("secret")

		stdErr := errors.New("standard error")
		log.Error("test error with std err", stdErr, nil)
		log.Error("test error with std err fields", stdErr, map[string]any{"key": "value"})

		appEx := exceptions.NewNotFoundException("entity not found", errors.New("db error"))
		log.Error("test error with exception", appEx, map[string]any{"secret": "hidden"})
	})
}
