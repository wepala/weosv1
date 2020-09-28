package domain

type BasicEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Meta    EventMeta   `json:"meta"`
	Version int         `json:"version"`
}

func (b *BasicEvent) GetID() string {
	return b.Meta.ID
}

func (b *BasicEvent) GetMetadata() EventMeta {
	return b.Meta
}

func (b *BasicEvent) GetPayload() interface{} {
	return b.Payload
}

func (b *BasicEvent) GetType() string {
	return b.Type
}

type EventMeta struct {
	ID          string `json:"id"`
	SequenceNo  int64  `json:"sequenceNo"`
	User        string `json:"user"`
	Application string `json:"application"`
	Account     string `json:"account"`
	Created     string `json:"created"`
}
