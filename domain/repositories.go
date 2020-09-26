package domain

type EventRepository interface {
	GetByAggregate(ID string) []*Event
	GetByAggregateAndSequenceRange(ID string, start int64, end int64) []*Event
	Save([]*Event) error
}
