package entities

type Entity interface {
	ValueObject
	GetID() string
}

type ValueObject interface {
	IsValid() bool
}

type EventSourcedEntity interface {
	Entity
	ApplyEvent(event *Event)
	ApplyEventHistory(event []*Event)
	NewChange(event *Event)
	GetNewChanges() []*Event
}
