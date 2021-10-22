package weos_test

import (
	"encoding/json"
	"os"
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

var clientID *redis.Client
var client *redis.Client

func TestMain(m *testing.M) {
	//setup redis to run in docker
	log.Infof("Started redis")
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "docker.redis.co/redis/redistestinstance:7.10.2",
		Name:         "redis7-mock",
		ExposedPorts: []string{"6379:6379/tcp", "6379:6379/tcp"},
		Env:          map[string]string{"discovery.type": "single-node"},
		WaitingFor:   wait.ForLog("started"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start redis container '%s'", err)
	}

	defer redisContainer.Terminate(ctx)

	req1 := testcontainers.ContainerRequest{
		Image:        "docker.redis.co/redis/redistestinstance2:7.10.2",
		Name:         "redis7-mock2",
		ExposedPorts: []string{"6379:6379/tcp", "6378:6378/tcp"},
		Env:          map[string]string{"discovery.type": "single-node"},
		WaitingFor:   wait.ForLog("started"),
	}
	redisContainer1, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req1,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start redis container '%s'", err)
	}

	defer redisContainer1.Terminate(ctx)

	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	clientID = redis.NewClient(&redis.Options{
		Addr: "localhost:6378",
	})

	if err != nil {
		log.Fatalf("error setting up redis client '%s'", err)
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

	eventRepository, err := weos.NewRedisEventRepository(client, clientID, log.New(), "accountID", "applicationID")
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

	entity := &weos.AggregateRoot{}
	entity.NewChange(mockEvent)

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "123"), entity)
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	var events []weos.RedisEvent
	results := clientID.Get(mockEvent.Meta.EntityID)
	if results.Err() != nil {
		t.Fatalf("error encountered getting event '%s'", err)
	}
	values := results.Val()

	err = json.Unmarshal([]byte(values), &events)

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
	results = client.Get(mockEvent.Meta.EntityID + ":" + mockEvent.Meta.EntityType + ":" + mockEvent.Meta.RootID)
	if results.Err() != nil {
		t.Fatalf("error encountered getting event '%s'", err)
	}
	values = results.Val()

	err = json.Unmarshal([]byte(values), &events)

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
