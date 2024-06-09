package messages

import (
    "encoding/json"
    "github.com/google/uuid"
    "time"
)

type Meta struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Version   uint8     `json:"version"`
}

type Message struct {
    Meta    Meta `json:"meta"`
    Payload any  `json:"payload"`
}

func createMeta(v uint8) Meta {
    return Meta{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        Version:   v,
    }
}

func NewMessage(v uint8, payload any) ([]byte, error) {
    m := Message{
        Meta:    createMeta(v),
        Payload: payload,
    }

    j, err := json.Marshal(m)

    if err != nil {
        return nil, err
    }

    return j, nil
}

func FromJSON(b []byte) (Message, error) {
    var m Message

    err := json.Unmarshal(b, &m)

    if err != nil {
        return Message{}, err
    }

    return m, nil
}
