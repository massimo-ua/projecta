package messages

type EventTopic = string

const (
	PurchaseConfirmed EventTopic = "purchase.confirmed"
	PurchaseRefunded  EventTopic = "purchase.refunded"
	StoreRegistered   EventTopic = "store.registered"
)
