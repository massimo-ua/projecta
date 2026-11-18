package logger

import (
	"errors"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"log/slog"
	"os"
	"strings"
)

type JsonLogger struct {
	logger *slog.Logger
}

func handleEmptyFields(fields map[string]any) map[string]any {
	if fields == nil {
		return make(map[string]any)
	}

	return fields
}

func (l *JsonLogger) Info(message string, fields map[string]any) {
	var details []any

	fields = handleEmptyFields(fields)

	for key, value := range fields {
		details = append(details, slog.Any(key, value))
	}

	l.logger.Info(message, details...)
}

func (l *JsonLogger) Error(message string, err error, fields map[string]any) {
	var details []any

	fields = handleEmptyFields(fields)

	for key, value := range fields {
		details = append(details, slog.Any(key, value))
	}

	var ex exceptions.Exception

	isException := errors.As(err, &ex)

	var errorCode string
	var errorMessage string

	if isException {
		errorCode = ex.Code
		errorMessage = ex.Message
	} else if err != nil {
		errorCode = exceptions.Internal
		errorMessage = err.Error()
	} else {
		errorCode = exceptions.Internal
		errorMessage = "unknown error"
	}

	details = append(details, slog.Any("code", errorCode), slog.Any("message", errorMessage))

	l.logger.Error(message, details...)
}

var defaultProtectedKeys = []string{
	"password",
	"token",
	"secret",
	"authorization",
	"refresh_token",
	"access_token",
	"db_uri",
	"jwt_secret",
}

func New(protectedKeys ...string) core.Logger {
	allProtected := make([]string, 0, len(defaultProtectedKeys)+len(protectedKeys))
	allProtected = append(allProtected, defaultProtectedKeys...)
	allProtected = append(allProtected, protectedKeys...)

	return &JsonLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Sanitize sensitive information
				for _, key := range allProtected {
					if strings.EqualFold(a.Key, key) {
						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue("***"),
						}
					}
				}

				return a
			},
		})),
	}
}
