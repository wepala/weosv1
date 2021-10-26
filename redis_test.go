package weos_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos"
	"golang.org/x/net/context"
)

var database *redis.Client

func TestMain(m *testing.M) {
	//setup redis to run in docker
	log.Infof("Started redis")

	database = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := database.Ping().Result()
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestNewRedisEvent(t *testing.T) {
	event := &weos.Event{
		ID:      "event1",
		Type:    "type2",
		Payload: nil,
		Meta: weos.EventMeta{
			EntityID:   "123sg",
			EntityType: "testing",
			SequenceNo: 987,
			User:       "user213",
			Module:     "app123",
			RootID:     "123sg",
			Created:    time.Now().Format(time.RFC3339Nano),
		},
	}
	redisEvent, err := weos.NewRedisEvent(event)
	if err != nil {
		t.Errorf("unexpected err creating a new redis event")
	}
	if event.ID != redisEvent.ID {
		t.Errorf("expected redis event id to be  '%s', got '%s'", event.ID, redisEvent.ID)
	}
	if event.Type != redisEvent.Type {
		t.Errorf("expected redis event type to be '%s', got '%s'", event.Type, redisEvent.Type)
	}
	if string(redisEvent.Payload) != "null" {
		t.Errorf("expected redis event to be nil")
	}
	if event.Meta.EntityID != redisEvent.Meta.EntityID {
		t.Errorf("expected redis entity id to be  '%s', got '%s'", event.Meta.EntityID, redisEvent.Meta.EntityID)
	}
	if event.Meta.EntityType != redisEvent.Meta.EntityType {
		t.Errorf("expected redis entity type to be  '%s', got '%s'", event.Meta.EntityType, redisEvent.Meta.EntityType)
	}
	if event.Meta.SequenceNo != redisEvent.Meta.SequenceNo {
		t.Errorf("expected redis sequence number to be  '%d', got '%d'", event.Meta.SequenceNo, redisEvent.Meta.SequenceNo)
	}
	if event.Meta.RootID != redisEvent.Meta.RootID {
		t.Errorf("expected redis root id to be  '%s', got '%s'", event.Meta.RootID, redisEvent.Meta.RootID)
	}
	if event.Meta.User != redisEvent.Meta.User {
		t.Errorf("expected redis user to be  '%s', got '%s'", event.Meta.User, redisEvent.Meta.User)
	}
	if event.Meta.Created != redisEvent.Meta.Created {
		t.Errorf("expected redis entity id to be  '%s', got '%s'", event.Meta.Created, redisEvent.Meta.Created)
	}
}

func TestPersist(t *testing.T) {

	eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	mockEvent := &weos.Event{
		ID:      ksuid.New().String(),
		Type:    "TEST_EVENT",
		Payload: nil,
		Meta: weos.EventMeta{
			EntityID:   "some id",
			EntityType: "SomeType",
			SequenceNo: 0,
			RootID:     "root123",
		},
		Version: 1,
	}

	//add an event handler
	eventHandlerCalled := 0
	eventRepository.AddSubscriber(func(ctx context.Context, event weos.Event) {
		eventHandlerCalled += 1
	})

	entity := &weos.AggregateRoot{BasicEntity: weos.BasicEntity{ID: "2635sgbd"}}
	entity.NewChange(mockEvent)

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "root123"), entity)
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	var events []weos.RedisEvent

	results := database.Get(mockEvent.Meta.EntityID + ":" + mockEvent.Meta.EntityType + ":" + mockEvent.Meta.RootID)
	if results.Err() != nil {
		t.Fatalf("error encountered getting event '%s'", err)
	}

	err = json.Unmarshal([]byte(results.Val()), &events)

	if err != nil {
		t.Fatalf("error encountered unmarshalling events '%s'", err)
	}
	if len(events) == 0 {
		t.Fatalf("unexpected error, no events found")
	}
	if events[0].ID != mockEvent.ID {
		t.Errorf("expected event id to be  '%s', got '%s'", mockEvent.ID, events[0].ID)
	}
	if events[0].Type != mockEvent.Type {
		t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
	}
	if events[0].Type != mockEvent.Type {
		t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
	}
	if string(events[0].Payload) != "null" {
		t.Errorf("expected event to be nil")
	}
	if events[0].Meta.EntityID != mockEvent.Meta.EntityID {
		t.Errorf("expected entity id to be  '%s', got '%s'", mockEvent.Meta.EntityID, events[0].Meta.EntityID)
	}
	if events[0].Meta.EntityID != mockEvent.Meta.EntityID {
		t.Errorf("expected entity type to be  '%s', got '%s'", mockEvent.Meta.EntityID, events[0].Meta.EntityID)
	}
	if events[0].Meta.User != mockEvent.Meta.User {
		t.Errorf("expected user to be  '%s', got '%s'", mockEvent.Meta.User, events[0].Meta.User)
	}
	if events[0].Meta.SequenceNo != mockEvent.Meta.SequenceNo {
		t.Errorf("expected event type to be  '%d', got '%d'", mockEvent.Meta.SequenceNo, events[0].Meta.SequenceNo)
	}
	if events[0].Version != mockEvent.Version {
		t.Errorf("expected event type to be  '%d', got '%d'", mockEvent.Version, events[0].Version)
	}
}

func TestGetByEntityAndAggregate(t *testing.T) {

	t.Run("get aggregate with 1 event ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "root123", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}
		entity := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Post", RootID: "root123", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Post", Description: "Lorem Ipsum"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 1,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Post", Description: "Finalizing Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 3,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		entity.NewChange(mockEvent)
		entity.NewChange(mockEvent2)
		entity.NewChange(mockEvent3)

		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity.RootID), entity)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}

		events, err := eventRepository.GetByEntityAndAggregate(entity.ID, entity.Type, entity.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity.ID, err)
		}

		if len(events) != 3 {
			t.Errorf("expected %d events got %d", 3, len(events))
		}

	})

	t.Run("get aggregate with 2 events with same type ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "root123", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}

		entity := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Post", RootID: "root123", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"}}}
		entity1 := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Post", RootID: "root123", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "iNfR0jYD9UbY"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Post", Description: "Lorem Ipsum"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 1,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Post", Description: "Finalizing Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_POST",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity1.ID,
				EntityType: entity1.Type,
				SequenceNo: 0,
				RootID:     entity1.RootID,
			},
			Version: 1,
		}

		entity.NewChange(mockEvent)
		entity.NewChange(mockEvent2)
		entity1.NewChange(mockEvent3)

		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity.RootID), entity)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}
		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity1.RootID), entity1)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}

		events, err := eventRepository.GetByEntityAndAggregate(entity.ID, entity.Type, entity.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity.ID, err)
		}

		if len(events) != 2 {
			t.Errorf("expected %d events got %d", 2, len(events))
		}

		events, err = eventRepository.GetByEntityAndAggregate(entity1.ID, entity1.Type, entity1.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity1.ID, err)
		}

		if len(events) != 1 {
			t.Errorf("expected %d events got %d", 1, len(events))
		}
	})
}
