package broker

import (
	"context"
	"errors"
	"sync"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type mockConnection struct {
	closeErr error
}

func (m *mockConnection) Close() error {
	return m.closeErr
}

func (m *mockConnection) Channel() (*amqp.Channel, error) {
	return nil, nil
}

type mockConnectionWithChannel struct {
	channelErr error
	closeErr   error
}

func (m *mockConnectionWithChannel) Channel() (*amqp.Channel, error) {
	if m.channelErr != nil {
		return nil, m.channelErr
	}
	return nil, nil
}

func (m *mockConnectionWithChannel) Close() error {
	return m.closeErr
}

type mockChannel struct {
	exchangeErr error
	publishErr  error
	queueErr    error
	bindErr     error
	consumeErr  error
	closeErr    error
	deliveries  chan amqp.Delivery
}

func (m *mockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return m.exchangeErr
}

func (m *mockChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return m.publishErr
}

func (m *mockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if m.queueErr != nil {
		return amqp.Queue{}, m.queueErr
	}
	return amqp.Queue{Name: "test_queue"}, nil
}

func (m *mockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return m.bindErr
}

func (m *mockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if m.consumeErr != nil {
		return nil, m.consumeErr
	}
	if m.deliveries == nil {
		m.deliveries = make(chan amqp.Delivery, 1)
	}
	return m.deliveries, nil
}

func (m *mockChannel) Close() error {
	return m.closeErr
}

func TestAMQPBroker(t *testing.T) {
	t.Run("NewAMQPBroker empty URL error", func(t *testing.T) {
		_, err := NewAMQPBroker("")
		if err == nil {
			t.Errorf("expected error for empty URL")
		}
	})

	t.Run("NewAMQPBroker invalid connection URL error", func(t *testing.T) {
		_, err := NewAMQPBroker("amqp://invalid-host:5672")
		if err == nil {
			t.Errorf("expected error for invalid host")
		}
	})

	t.Run("NewAMQPBroker success and channel error with mocked dialAMQP", func(t *testing.T) {
		oldDial := dialAMQP
		defer func() { dialAMQP = oldDial }()

		// Mock dialAMQP returning error
		dialAMQP = func(url string) (amqpConnection, error) {
			return nil, errors.New("mock dial error")
		}

		_, err := NewAMQPBroker("amqp://localhost:5672")
		if err == nil {
			t.Errorf("expected mock dial error")
		}

		// Mock dialAMQP returning connection with channel error
		dialAMQP = func(url string) (amqpConnection, error) {
			return &mockConnectionWithChannel{channelErr: errors.New("mock channel error")}, nil
		}

		_, err = NewAMQPBroker("amqp://localhost:5672")
		if err == nil {
			t.Errorf("expected mock channel error")
		}

		// Mock dialAMQP returning success connection
		dialAMQP = func(url string) (amqpConnection, error) {
			return &mockConnectionWithChannel{}, nil
		}

		b, err := NewAMQPBroker("amqp://localhost:5672")
		if err != nil || b == nil {
			t.Fatalf("unexpected error for mock success: %v", err)
		}
	})

	t.Run("Publish success and error branches", func(t *testing.T) {
		ch := &mockChannel{}
		b := &AMQPBroker{ch: ch}

		err := b.Publish(context.Background(), "test.topic", []byte(`{"key":"val"}`))
		if err != nil {
			t.Fatalf("unexpected publish error: %v", err)
		}

		ch.exchangeErr = errors.New("exchange declare error")
		err = b.Publish(context.Background(), "test.topic", []byte("{}"))
		if err == nil {
			t.Errorf("expected error when ExchangeDeclare fails")
		}

		ch.exchangeErr = nil
		ch.publishErr = errors.New("publish error")
		err = b.Publish(context.Background(), "test.topic", []byte("{}"))
		if err == nil {
			t.Errorf("expected error when PublishWithContext fails")
		}
	})

	t.Run("Subscribe success and error branches", func(t *testing.T) {
		ch := &mockChannel{deliveries: make(chan amqp.Delivery, 1)}
		b := &AMQPBroker{ch: ch}

		var wg sync.WaitGroup
		var receivedMsg []byte
		wg.Add(1)

		handler := func(msg []byte) {
			receivedMsg = msg
			wg.Done()
		}

		cancel, err := b.Subscribe(context.Background(), "test.topic", handler)
		if err != nil || cancel == nil {
			t.Fatalf("unexpected subscribe error: %v", err)
		}

		ch.deliveries <- amqp.Delivery{Body: []byte("hello amqp")}
		wg.Wait()
		cancel()

		if string(receivedMsg) != "hello amqp" {
			t.Errorf("expected 'hello amqp', got '%s'", string(receivedMsg))
		}

		// ExchangeDeclare error
		ch.exchangeErr = errors.New("exchange err")
		_, err = b.Subscribe(context.Background(), "topic", handler)
		if err == nil {
			t.Errorf("expected exchange error")
		}
		ch.exchangeErr = nil

		// QueueDeclare error
		ch.queueErr = errors.New("queue err")
		_, err = b.Subscribe(context.Background(), "topic", handler)
		if err == nil {
			t.Errorf("expected queue error")
		}
		ch.queueErr = nil

		// QueueBind error
		ch.bindErr = errors.New("bind err")
		_, err = b.Subscribe(context.Background(), "topic", handler)
		if err == nil {
			t.Errorf("expected bind error")
		}
		ch.bindErr = nil

		// Consume error
		ch.consumeErr = errors.New("consume err")
		_, err = b.Subscribe(context.Background(), "topic", handler)
		if err == nil {
			t.Errorf("expected consume error")
		}
	})

	t.Run("Close with and without errors", func(t *testing.T) {
		conn := &mockConnection{}
		ch := &mockChannel{}
		b := &AMQPBroker{conn: conn, ch: ch}

		// Success close
		b.Close()

		// Error close
		connErr := &mockConnection{closeErr: errors.New("conn close err")}
		chErr := &mockChannel{closeErr: errors.New("ch close err")}
		bErr := &AMQPBroker{conn: connErr, ch: chErr}
		bErr.Close()
	})
}
