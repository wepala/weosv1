package weos

import "context"

//go:generate moq -out mocks_test.go -pkg weos_test . EventRepository
type WeOSEntity interface {
	Entity
	GetUser() User
	SetUser(User)
}
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
	GetNewChanges() []Entity
}

type Reducer func(initialState Entity, event Event, next Reducer) Entity

type Repository interface {
	Persist(entities []Entity) error
	Remove(entities []Entity) error
}

type UnitOfWorkRepository interface {
	Flush() error
}

type EventRepository interface {
	UnitOfWorkRepository
	Datastore
	Persist(entity AggregateInterface, meta EventMeta) error
	GetByAggregate(ID string) ([]*Event, error)
	GetByAggregateAndType(ID string, entityType string) ([]*Event, error)
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*Event, error)
	AddSubscriber(handler EventHandler)
	GetSubscribers() ([]EventHandler, error)
}

type Datastore interface {
	Migrate(ctx context.Context) error
}

type Projection interface {
	Datastore
	GetEventHandler() EventHandler
}
