package repositories

import "github.com/wepala/weos/domain"

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
}
