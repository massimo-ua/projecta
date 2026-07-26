package broker

import (
    "context"
    "fmt"
    amqp "github.com/rabbitmq/amqp091-go"
)

type amqpConnection interface {
	Channel() (*amqp.Channel, error)
	Close() error
}

type amqpChannel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type AMQPBroker struct {
    conn amqpConnection
    ch   amqpChannel
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

var dialAMQP = func(url string) (amqpConnection, error) {
	return amqp.Dial(url)
}

func NewAMQPBroker(connectionURL string) (*AMQPBroker, error) {
	if connectionURL == "" {
		return nil, fmt.Errorf("broker connection url is empty")
	}

	conn, err := dialAMQP(connectionURL)

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
