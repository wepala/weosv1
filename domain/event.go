package domain

import (
	"time"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Meta    EventMeta   `json:"meta"`
	Version int         `json:"version"`
}

func NewBasicEvent(eventType string, id string, payload interface{}, creatorID string) *Event {
	return &Event{
		Type:    eventType,
		Payload: payload,
		Meta: EventMeta{
			ID:      id,
			User:    creatorID,
			Created: time.Now().Format(time.RFC3339Nano),
		},
	}
}

type EventMeta struct {
	ID          string `json:"id"`
	SequenceNo  int64  `json:"sequenceNo"`
	User        string `json:"user"`
	Application string `json:"application"`
	Account     string `json:"account"`
	Created     string `json:"created"`
}
