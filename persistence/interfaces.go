package persistence

//go:generate moq -out persistence_mocks_test.go -pkg persistence_test . EventRepository Projection

import (
	"context"
	"github.com/wepala/weos/domain"
)

type Repository interface {
	Persist(entities []domain.Entity) error
	Remove(entities []domain.Entity) error
}

type UnitOfWorkRepository interface {
	Flush() error
}

type EventRepository interface {
	Repository
	UnitOfWorkRepository
	GetByAggregate(ID string) ([]*domain.Event, error)
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*domain.Event, error)
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
