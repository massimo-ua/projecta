package websocket

type MessageDTO struct {
	MessageType int
	Payload     []byte
}

type PayloadDTO struct {
	Type  string `json:"type"`
	Token string `json:"token,omitempty"`
	Data  string `json:"data,omitempty"`
}
