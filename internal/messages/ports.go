package messages

import (
    "context"
)

type MessageBroker interface {
    Publish(ctx context.Context, topic string, message []byte) error
    Subscribe(ctx context.Context, topic string, handler func(message []byte)) (context.CancelFunc, error)
}
