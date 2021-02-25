package domain

import (
	"encoding/json"
	"github.com/segmentio/ksuid"
	"github.com/wepala/weos/errors"
	"time"
)

type Event struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Meta    EventMeta       `json:"meta"`
	Version int             `json:"version"`
	errors  []error
}

var NewBasicEvent = func(eventType string, entityID string, entityType string, payload interface{}) (*Event, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.NewDomainError("Unable to marshal event payload", eventType, entityID, err)
	}
	return &Event{
		ID:      ksuid.New().String(),
		Type:    eventType,
		Payload: payloadBytes,
		Version: 1,
		Meta: EventMeta{
			EntityID:   entityID,
			EntityType: entityType,
			Created:    time.Now().Format(time.RFC3339Nano),
		},
	}, nil
}

var NewVersionEvent = func(eventType string, entityID string, payload interface{}, version int) (*Event, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.NewDomainError("Unable to marshal event payload", eventType, entityID, err)
	}
	return &Event{
		ID:      ksuid.New().String(),
		Type:    eventType,
		Payload: payloadBytes,
		Version: version,
		Meta: EventMeta{
			EntityID: entityID,
			Created:  time.Now().Format(time.RFC3339Nano),
		},
	}, nil
}

type EventMeta struct {
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
	SequenceNo int64  `json:"sequenceNo"`
	User       string `json:"user"`
	Module     string `json:"module"`
	Account    string `json:"account"`
	Group      string `json:"group"`
	Created    string `json:"created"`
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

	if e.Meta.EntityType == "" {
		e.AddError(errors.NewDomainError("all domain events must have an entity type", "Event", e.Meta.EntityID, nil))
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
