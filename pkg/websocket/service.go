package websocket

import (
	"encoding/json"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type AppAdapter struct{}

type IncomingMessage struct {
	MessageType int
	Payload     []byte
}

type OutgoingMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type PingMessage struct {
	Type string `json:"type"`
}

func NewAppAdapter() *AppAdapter {
	return &AppAdapter{}
}

func (a *AppAdapter) Handle(m IncomingMessage) ([]byte, error) {
	var pingMessage PingMessage
	err := json.Unmarshal(m.Payload, &pingMessage)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to decode message", err)
	}

	response, err := json.Marshal(OutgoingMessage{
		Type:    "ping",
		Message: "pong",
	})

	if err != nil {
		return nil, exceptions.NewInternalException("failed to encode response", err)
	}

	return response, nil
}
