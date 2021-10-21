package weos

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/segmentio/ksuid"
)

type TestEvent struct {
	ID   string `dynamo:"event_id,hash"`
	Name string `dynamo:"name"`
}

type DynamoEventService struct {
	eventTable dynamo.Table
}

func NewEventService(DynamoEventName string) (*DynamoEventService, error) {
	dynamoTable, err := newDynamoTable(DynamoEventName, "")
	if err != nil {
		return nil, err
	}
	return &DynamoEventService{
		eventTable: dynamoTable,
	}, nil
}

func newDynamoTable(tableName, endpoint string) (dynamo.Table, error) {
	if tableName == "" {
		return dynamo.Table{}, fmt.Errorf("you must supply a table name")
	}
	cfg := aws.Config{}
	cfg.Region = aws.String("eu-west-2")
	cfg.Credentials = credentials.AnonymousCredentials
	if endpoint != "" {
		cfg.Endpoint = aws.String("http://localhost:8000")
	}
	sess := session.Must(session.NewSession())
	db := dynamo.New(sess, &cfg)
	table := db.Table(tableName)
	return table, nil
}

func (d *DynamoEventService) CreateEvent(event *TestEvent) error {
	event.ID = ksuid.New().String()
	event.Name = "Testing Dynamo Create"
	return d.eventTable.Put(event).Run()
}

func (d *DynamoEventService) GetEvent(event *TestEvent) error {
	return d.eventTable.Get("event_id", event.ID).One(event)
}
