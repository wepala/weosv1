package weos_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wepala/weos"
	"golang.org/x/net/context"
)

var database *redis.Client

func TestMain(m *testing.M) {
	//setup redis to run in docker
	log.Infof("Started redis")
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis",
		Name:         "redis-mock",
		ExposedPorts: []string{"6379:6379/tcp"},
		Env:          map[string]string{"REDIS_DB_URL": "redis:6379", "REDIS_DB_PASSWORD": "", "REDIS_DB": "0"},
		WaitingFor:   wait.ForLog("started"),
	}
	rContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start elastic search container '%s'", err)
	}

	defer rContainer.Terminate(ctx)

	//get the endpoint that the container was run on
	var endpoint string
	endpoint, err = rContainer.Host(ctx) //didn't use the endpoint call because it returns "localhost" which the client doesn't seem to like
	if err != nil {
		log.Fatalf("error setting up redis '%s'", err)
	}
	cport, err := rContainer.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("error setting up redis '%s'", err)
	}
	rEndpoint := endpoint + ":" + cport.Port()

	database = redis.NewClient(&redis.Options{
		Addr:     rEndpoint,
		Password: "",
		DB:       0,
	})
	pong, err := database.Ping().Result()
	if err != nil {
		panic(err)
	}

	if pong == "" {
		panic("no pong received")
	}

	code := m.Run()
	os.Exit(code)
}

func TestRedis_NewRedisEvent(t *testing.T) {
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

func TestRedis_Persist(t *testing.T) {

	eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "restaurant", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	mockEvent := &weos.Event{
		ID:      ksuid.New().String(),
		Type:    "UPDATE_FOOD",
		Payload: nil,
		Meta: weos.EventMeta{
			EntityID:   "chicken",
			EntityType: "food",
			SequenceNo: 0,
			RootID:     "restaurant",
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

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "resturant"), entity)
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

func TestRedis_GetByEntityAndAggregate(t *testing.T) {

	t.Run("get aggregate with 1 event ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "store", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}
		entity := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Item", RootID: "store", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_ITEM",
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
		}{Title: "Updated First Item", Description: "Shiny pearly necklace"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_ITEM",
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
		}{Title: "Updated First Item", Description: "Finalizing Item"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_ITEM",
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
		if events[0].ID != mockEvent.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent.ID, events[0].ID)
		}
		if events[0].Type != mockEvent.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
		}
		if events[1].ID != mockEvent2.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent2.ID, events[1].ID)
		}
		if events[1].Type != mockEvent2.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent2.Type, events[1].Type)
		}
		if events[2].ID != mockEvent3.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent3.ID, events[2].ID)
		}
		if events[2].Type != mockEvent3.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Type, events[2].Type)
		}

	})

	t.Run("get aggregate with 2 events with same type ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "shop", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}

		entity := &struct {
			weos.AggregateRoot
			Type string `json:"type"`
		}{Type: "Snacks", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "another snack"}}}
		entity1 := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Snacks", RootID: "shop", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "some snack"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Snack"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_SNACK",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Snack", Description: "Lorem Ipsum"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_SNACK",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 1,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Create Second Snack", Description: "Snack Snack Snack"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_SNACK",
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

		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "shop"), entity)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}
		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity1.RootID), entity1)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}

		events, err := eventRepository.GetByEntityAndAggregate(entity.ID, entity.Type, "shop")
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity.ID, err)
		}

		if len(events) != 2 {
			t.Errorf("expected %d events got %d", 2, len(events))
		}
		if events[0].ID != mockEvent.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent.ID, events[0].ID)
		}
		if events[0].Type != mockEvent.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
		}
		if events[0].Meta.RootID != "shop" {
			t.Errorf("expected event type to be  '%s', got '%s'", "shop", events[0].Type)
		}
		if events[1].ID != mockEvent2.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent2.ID, events[1].ID)
		}
		if events[1].Type != mockEvent2.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[1].Type)
		}
		if events[1].Meta.RootID != "shop" {
			t.Errorf("expected event type to be  '%s', got '%s'", "shop", events[1].Type)
		}

		events, err = eventRepository.GetByEntityAndAggregate(entity1.ID, entity1.Type, entity1.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity1.ID, err)
		}

		if len(events) != 1 {
			t.Errorf("expected %d events got %d", 1, len(events))
		}
		if events[0].ID != mockEvent3.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent3.ID, events[0].ID)
		}
		if events[0].Type != mockEvent3.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Type, events[0].Type)
		}
		if events[0].Meta.RootID != mockEvent3.Meta.RootID {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Meta.RootID, events[0].Type)
		}
	})
}

func TestRedis_BatchPersist(t *testing.T) {

	eventRepository, err := weos.NewRedisEventRepository(database, log.New(), "root123", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	generateEvents := make([]*weos.Event, 20000)
	entity := &struct {
		weos.AggregateRoot
		Type   string `json:"type"`
		RootID string `json:"root_id"`
	}{Type: "Post", RootID: "root123", AggregateRoot: weos.AggregateRoot{
		BasicEntity: weos.BasicEntity{ID: "batch id"}}}

	for i := 0; i < 20000; i++ {

		currValue := strconv.Itoa(i)
		currEvent := "TEST_EVENT "
		currEvent += currValue

		generateEvents[i] = &weos.Event{
			ID:      ksuid.New().String(),
			Type:    currEvent,
			Payload: nil,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		entity.NewChange(generateEvents[i])
	}

	//add an event handler
	eventHandlerCalled := 0
	eventRepository.AddSubscriber(func(ctx context.Context, event weos.Event) {
		eventHandlerCalled += 1
	})

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity.RootID), entity)
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	if eventHandlerCalled != 20000 {
		t.Errorf("expected event handlers to be called %d time, called %d times", 20000, eventHandlerCalled)
	}
	var events []weos.RedisEvent

	results := database.Get(entity.ID + ":" + entity.Type + ":" + entity.RootID)
	if results.Err() != nil {
		t.Fatalf("error encountered getting event '%s'", err)
	}

	err = json.Unmarshal([]byte(results.Val()), &events)

	if err != nil {
		t.Fatalf("error encountered unmarshalling events '%s'", err)
	}
	if len(events) != 20000 {
		t.Fatalf("expected events to have %d got %d", 20000, len(events))
	}

}
