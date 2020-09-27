package entities

type Entity interface {
	ValueObject
	GetID() string
}

type ValueObject interface {
	IsValid() bool
	AddError(err error)
	GetErrors() []error
}

type EventSourcedEntity interface {
	Entity
	ApplyEvent(event *Event)
	ApplyEventHistory(event []*Event)
	NewChange(event *Event)
	GetNewChanges() []*Event
}
