package repositories

import "github.com/wepala/weos/domain"

type Repository interface {
	Persist(entities []domain.Entity) error
}

type EventRepository interface {
	Persist(entities []domain.Event) error
	GetByAggregate(ID string) ([]domain.Event, error)
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]domain.Event, error)
	AddSubscriber(handler EventHandler)
}
