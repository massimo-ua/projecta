package websocket

import (
	"context"
	"encoding/json"
)

type AppAdapter interface {
	Handle(ctx context.Context, eventType string, data string) ([]byte, error)
}

type PayloadType = string

const (
	Ping  PayloadType = "ping"
	Error PayloadType = "error"
)

type AppAdapterImpl struct{}

var (
	pongResponse            = PayloadDTO{Type: Ping, Data: "pong"}
	noActiveHandlerResponse = PayloadDTO{Type: Error, Data: "no active handler"}
)

func NewAppAdapter() *AppAdapterImpl {
	return &AppAdapterImpl{}
}

func (a *AppAdapterImpl) pong() ([]byte, error) {
	return json.Marshal(pongResponse)
}

func (a *AppAdapterImpl) Handle(ctx context.Context, eventType string, data string) ([]byte, error) {
	switch eventType {
	case Ping:
		return a.pong()
	default:
		return json.Marshal(noActiveHandlerResponse)
	}
}
