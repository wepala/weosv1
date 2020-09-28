package domain

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Meta    *EventMeta  `json:"meta"`
	Version int         `json:"version"`
}

type EventMeta struct {
	ID          string `json:"id"`
	SequenceNo  int64  `json:"sequenceNo"`
	User        string `json:"user"`
	Application string `json:"application"`
	Account     string `json:"account"`
	Created     string `json:"created"`
}
