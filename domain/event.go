package domain

import (
	"github.com/segmentio/ksuid"
	"github.com/wepala/weos/errors"
	"time"
)

type Event struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Meta    EventMeta   `json:"meta"`
	Version int         `json:"version"`
	errors  []error
}

var NewBasicEvent = func(eventType string, entityID string, payload interface{}, creatorID string) *Event {
	return &Event{
		ID:      ksuid.New().String(),
		Type:    eventType,
		Payload: payload,
		Version: 1,
		Meta: EventMeta{
			EntityID: entityID,
			User:     creatorID,
			Created:  time.Now().Format(time.RFC3339Nano),
		},
	}
}

type EventMeta struct {
	EntityID    string `json:"entity_id"`
	SequenceNo  int64  `json:"sequenceNo"`
	User        string `json:"user"`
	Application string `json:"application"`
	Account     string `json:"account"`
	Created     string `json:"created"`
}

func (e *Event) IsValid() bool {
	if e.ID == "" {
		e.AddError(errors.NewDomainError("all events must have an id", "Event", e.Meta.EntityID, nil))
		return false
	}

	if e.Meta.EntityID == "" {
		e.AddError(errors.NewDomainError("all domain events must be associated with an entity", "Event", e.Meta.EntityID, nil))
		return false
	}

	if e.Version == 0 {
		e.AddError(errors.NewDomainError("all domain events must have a version no.", "Event", e.Meta.EntityID, nil))
		return false
	}

	if e.Type == "" {
		e.AddError(errors.NewDomainError("all domain events must have a type", "Event", e.Meta.EntityID, nil))
		return false
	}

	return true
}

func (e *Event) AddError(err error) {
	e.errors = append(e.errors, err)
}

func (e *Event) GetErrors() []error {
	return e.errors
}

func (e *Event) GetID() string {
	return e.ID
}
