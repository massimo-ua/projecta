package core

type Logger interface {
	Info(message string, fields map[string]any)
	Error(message string, err error, fields map[string]any)
}
