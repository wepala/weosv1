package domain

import "github.com/wepala/weos/domain/entities"

type EventRepository interface {
	GetByAggregate(ID string) []*entities.Event
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) []*entities.Event
	Save([]*entities.Event) error
}
