package domain

//go:generate moq -out mocks_test.go -pkg domain_test . EventRepository

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
	NewChange(event *Event)
	GetNewChanges() []*Event
}

type EventRepository interface {
	GetByAggregate(ID string) []Event
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) []Event
	Save([]Event) error
}

type Reducer func(initialState Entity, event Event, next Reducer) Entity
