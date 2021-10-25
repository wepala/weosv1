package weos_test

import (
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos"
	"golang.org/x/net/context"
)

var dynamoDB *dynamodb.DynamoDB

func TestDynamo_CreateTable(t *testing.T) {

	err := weos.CreateTable()
	if err != nil {
		t.Fatalf("error creating table '%s'", err)
	}
}

func TestDynamo_AddEvent(t *testing.T) {

	eventRepository, err := weos.NewBasicEventRepositoryDynamo(dynamoDB, log.New(), true, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	generateEvents := make([]*weos.Event, 2000)
	entity := &weos.AggregateRoot{}

	for i := 0; i < 2000; i++ {

		currValue := strconv.Itoa(i)

		currEvent := "TEST_EVENT "
		currID := "batch id"
		currType := "SomeType"

		currEvent += currValue

		generateEvents[i] = &weos.Event{
			ID:      ksuid.New().String(),
			Type:    currEvent,
			Payload: nil,
			Meta: weos.EventMeta{
				EntityID:   currID,
				EntityType: currType,
				SequenceNo: 0,
			},
			Version: 1,
		}

		entity.NewChange(generateEvents[i])
	}

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "12345"), entity)
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}

	//Check for events
}

/*func TestDynamo_AddEvent(t *testing.T) {

	testEvent := weos.TestEvent{
		ID:     "3",
		Name:   "Test Event 3",
		Random: "12345",
	}

	err := weos.PutEvent(testEvent)
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}
}

func TestDynamo_GetEvent(t *testing.T) {

	testEvent, err := weos.GetEvent("2")
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}

	if testEvent.ID != "2" {
		t.Fatalf("Expected test event ID to be '%s', got '%s'", "2", testEvent.ID)
	}
	if testEvent.Name != "Test Event 2" {
		t.Fatalf("Expected test event name to be '%s', got '%s'", "Test Event 2", testEvent.Name)
	}

}*/
