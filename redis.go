package weos

import (
	"encoding/json"

	"github.com/go-redis/redis"
	"golang.org/x/net/context"
	"gorm.io/datatypes"
)

type RedisEvent struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Payload datatypes.JSON `json:"payload"`
	Meta    EventMeta      `json:"meta"`
	Version int            `json:"version"`
}

type EventRepositoryRedis struct {
	logger          Log
	clientID        *redis.Client //database with the key as entity id
	client          *redis.Client // database with a composite key of entity id,entity type and rootid
	index           string
	eventDispatcher EventDisptacher
	AccountID       string
	ApplicationID   string
	GroupID         string
	UserID          string
}

//NewRedisEvent converts a domain event to something that is a bit easier for redis to work with
func NewRedisEvent(event *Event) (RedisEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return RedisEvent{}, err
	}

	return RedisEvent{
		ID:      event.ID,
		Type:    event.Type,
		Payload: payload,
		Meta:    event.Meta,
		Version: event.Version,
	}, nil

}

//adds events to both database
func (r *EventRepositoryRedis) Persist(ctxt context.Context, entity AggregateInterface) error {
	var redisEvents []RedisEvent
	entities := entity.GetNewChanges()

	for _, entity := range entities {
		event := entity.(*Event)
		//let's fill in meta data if it's not already in the object
		if event.Meta.User == "" {
			event.Meta.User = GetUser(ctxt)
		}
		if event.Meta.RootID == "" {
			event.Meta.RootID = GetAccount(ctxt)
		}
		if event.Meta.Module == "" {
			event.Meta.Module = r.ApplicationID
		}
		if event.Meta.Group == "" {
			event.Meta.Group = r.GroupID
		}
		if !event.IsValid() {
			for _, terr := range event.GetErrors() {
				r.logger.Errorf("error encountered persisting entity '%s', '%s'", event.Meta.EntityID, terr)
			}

			return event.GetErrors()[0]
		}

		redisEvent, err := NewRedisEvent(event)
		if err != nil {
			return err
		}

		redisEvents = append(redisEvents, redisEvent)
	}
	marshalEvents, err := json.Marshal(redisEvents)
	if err != nil {
		return err
	}
	status := r.clientID.Set(redisEvents[0].Meta.EntityID, marshalEvents, 0)
	if status.Err() != nil {
		return status.Err()
	}

	status = r.client.Set(redisEvents[0].Meta.EntityID+":"+redisEvents[0].Meta.EntityType+":"+redisEvents[0].Meta.RootID, marshalEvents, 0)
	if status.Err() != nil {
		return status.Err()
	}

	//call persist on the aggregate root to clear the new changes array
	entity.Persist()

	for _, entity := range entities {
		r.eventDispatcher.Dispatch(ctxt, *entity.(*Event))
	}
	return nil
}

//GetByAggregate get events for a entity id
func (r *EventRepositoryRedis) GetByAggregate(entityID string) ([]*Event, error) {
	var events []RedisEvent
	results := r.clientID.Get(entityID)
	if results.Err() != nil {
		return nil, results.Err()
	}
	values := results.Val()

	err := json.Unmarshal([]byte(values), &events)

	if err != nil {
		return nil, err
	}

	var tevents []*Event

	for _, event := range events {
		tevents = append(tevents, &Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta:    event.Meta,
			Version: event.Version,
		})
	}
	return tevents, nil

}

//GetByEntityAndAggregate gets events for a entity id, entity type and rootid
func (r *EventRepositoryRedis) GetByEntityAndAggregate(entityID string, entityType string, rootID string) ([]*Event, error) {
	var events []RedisEvent
	results := r.client.Get(entityID + ":" + entityType + ":" + rootID)
	if results.Err() != nil {
		return nil, results.Err()
	}
	values := results.Val()

	err := json.Unmarshal([]byte(values), &events)

	if err != nil {
		return nil, err
	}

	var tevents []*Event

	for _, event := range events {
		tevents = append(tevents, &Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta:    event.Meta,
			Version: event.Version,
		})
	}
	return tevents, nil
}

//AddSubscriber Allows you to add a handler that is triggered when events are dispatched
func (r *EventRepositoryRedis) AddSubscriber(handler EventHandler) {
	r.eventDispatcher.AddSubscriber(handler)
}

//GetSubscribers Get the current list of event subscribers
func (r *EventRepositoryRedis) GetSubscribers() ([]EventHandler, error) {
	return r.eventDispatcher.GetSubscribers(), nil
}

func (r *EventRepositoryRedis) Flush() error {
	return nil
}

//GetAggregateSequenceNumber gets the latest sequence number for the aggregate entity
func (r *EventRepositoryRedis) GetAggregateSequenceNumber(ID string) (int64, error) {
	return 0, nil
}

func (r *EventRepositoryRedis) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*Event, error) {
	return nil, nil
}

func (r *EventRepositoryRedis) GetByAggregateAndType(ID string, entityType string) ([]*Event, error) {
	return nil, nil
}

func (r *EventRepositoryRedis) Migrate(ctx context.Context) error {
	return nil
}

func NewRedisEventRepository(client *redis.Client, clientID *redis.Client, logger Log, accountID string, applicationID string) (EventRepository, error) {

	return &EventRepositoryRedis{clientID: clientID, client: client, logger: logger, AccountID: accountID, ApplicationID: applicationID}, nil
}
