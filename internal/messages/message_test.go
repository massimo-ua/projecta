package messages_test

import (
	"testing"

	"gitlab.com/massimo-ua/projecta/internal/messages"
)

type samplePayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestMessages(t *testing.T) {
	t.Run("NewMessage and FromJSON roundtrip", func(t *testing.T) {
		payload := samplePayload{Name: "test", Value: 42}
		bytes, err := messages.NewMessage(1, payload)
		if err != nil {
			t.Fatalf("unexpected error creating message: %v", err)
		}
		if len(bytes) == 0 {
			t.Fatalf("expected non-empty json bytes")
		}

		msg, err := messages.FromJSON(bytes)
		if err != nil {
			t.Fatalf("unexpected error parsing JSON: %v", err)
		}
		if msg.Meta.Version != 1 {
			t.Errorf("expected version 1, got %d", msg.Meta.Version)
		}
		if msg.Meta.ID.String() == "" {
			t.Errorf("expected non-empty meta ID")
		}
	})

	t.Run("NewMessage JSON marshal error", func(t *testing.T) {
		unmarshalable := make(chan int)
		_, err := messages.NewMessage(1, unmarshalable)
		if err == nil {
			t.Errorf("expected error for unmarshalable payload")
		}
	})

	t.Run("FromJSON invalid json error", func(t *testing.T) {
		_, err := messages.FromJSON([]byte("invalid json"))
		if err == nil {
			t.Errorf("expected error for invalid json bytes")
		}
	})
}
