package domain

import "encoding/json"

type Entity interface {
	ValueObject
	GetID () string
}

type ValueObject interface {
	IsValid() bool
}

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Meta    *EventMeta      `json:"meta"`
	Version int `json:"version"`
}

type EventMeta struct {
	ID          string `json:"id"`
	SequenceNo  int64  `json:"sequenceNo"`
	User        string `json:"user"`
	Application string `json:"application"`
	Account     string `json:"account"`
	Created     string `json:"created"`
}

//TODO add validation to event data

type EventSourcedEntity interface {
	Entity
	ApplyEvent (event *Event)
	ApplyEventHistory (event []*Event)
	NewChange(event *Event)
	GetNewChanges() []*Event
}

//AggregateRoot base struct for microservices to use
type AggregateRoot struct {
	ID string
	newEvents []*Event
}

func (w *AggregateRoot) IsValid() bool {
	return w.ID != ""
}

func (w *AggregateRoot) GetID() string {
	return w.ID
}

func (w *AggregateRoot) NewChange(event *Event) {
	w.newEvents = append(w.newEvents,event)
}

func (w *AggregateRoot) GetNewChanges() []*Event {
	return w.newEvents
}

func (w *AggregateRoot) ApplyEvent(event *Event) {

}

func (w *AggregateRoot) ApplyEventHistory(event []*Event) {
	for _,event := range event {
		w.ApplyEvent(event)
	}
}



