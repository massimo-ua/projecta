package broker

import (
    "context"
    "fmt"
    amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPBroker struct {
    conn *amqp.Connection
    ch   *amqp.Channel
}

func (b *AMQPBroker) Publish(ctx context.Context, topic string, message []byte) error {
    err := b.ch.ExchangeDeclare(
        topic,
        "fanout",
        true,
        false,
        false,
        false,
        nil,
    )

    if err != nil {
        return fmt.Errorf("failed to declare an exchange: %s", err.Error())
    }

    err = b.ch.PublishWithContext(
        ctx,
        topic,
        "",
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
        })

    if err != nil {
        return fmt.Errorf("failed to publish a message: %s", err.Error())
    }

    return nil
}

func (b *AMQPBroker) Subscribe(ctx context.Context, topic string, handler func(message []byte)) (context.CancelFunc, error) {
    subCtx, cancel := context.WithCancel(ctx)

    err := b.ch.ExchangeDeclare(
        topic,
        "fanout",
        true,
        false,
        false,
        false,
        nil,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to declare an exchange: %s", err.Error())
    }

    q, err := b.ch.QueueDeclare(
        "",
        false,
        false,
        true,
        false,
        nil,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to declare a queue: %s", err.Error())
    }

    err = b.ch.QueueBind(
        q.Name,
        "",
        topic,
        false,
        nil,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to bind a queue: %s", err.Error())
    }

    msgs, err := b.ch.Consume(
        q.Name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to register a consumer: %s", err.Error())
    }

    go func() {
        for {
            select {
            case <-subCtx.Done():
                return
            case msg := <-msgs:
                handler(msg.Body)
            }
        }
    }()

    return cancel, nil
}

func (b *AMQPBroker) Close() {
    err := b.ch.Close()

    if err != nil {
        fmt.Printf("failed to close channel: %s", err.Error())
    }

    err = b.conn.Close()

    if err != nil {
        fmt.Printf("failed to close connection: %s", err.Error())
    }
}

func NewAMQPBroker(connectionURL string) (*AMQPBroker, error) {
    if connectionURL == "" {
        return nil, fmt.Errorf("broker connection url is empty")
    }

    conn, err := amqp.Dial(connectionURL)

    if err != nil {
        return nil, fmt.Errorf("failed to connect to broker: %s", err.Error())
    }

    ch, err := conn.Channel()

    if err != nil {
        return nil, fmt.Errorf("failed to open a channel: %s", err.Error())
    }

    return &AMQPBroker{
        ch:   ch,
        conn: conn,
    }, nil
}
