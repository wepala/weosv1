package weos

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"golang.org/x/net/context"
	"gorm.io/datatypes"
)

/*COMMANDS -
docker-compose up

For AWS CLI cmds, add --endpoint-url http://localhost:8000 at the end to hit local dynamo

aws dynamodb list-tables --endpoint-url http://localhost:8000
aws dynamodb scan --table-name Events --endpoint-url http://localhost:8000
aws dynamodb delete-table --table-name Events --endpoint-url http://localhost:8000
*/

/*
type EventRepositoryDynamo struct {
	DB              *dynamodb.DynamoDB
	eventDispatcher EventDisptacher
	logger          Log
	unitOfWork      bool
	AccountID       string
	ApplicationID   string
	GroupID         string
	UserID          string
}

type DynamoEvent struct {
	ID            string `dynamo:"item_id,hash"`
	EntityID      string `dynamo:"entity_id"`
	EntityType    string `dynamo:"entity_type"`
	Payload       datatypes.JSON
	Type          string `dynamo:"type"`
	RootID        string `dynamo:"root_id"`
	ApplicationID string `dynamo:"application_id"`
	User          string `dynamo:"user"`
	SequenceNo    int64
}*/

/*type TestEvent struct {
	ID     string `dynamo:"event_id,hash"`
	Name   string `dynamo:"name"`
	Random string `dynamo:"Random"`
}*/

const TABLE_NAME = "Events"

type EventRepositoryDynamo struct {
	DB              *dynamodb.DynamoDB
	eventDispatcher EventDisptacher
	logger          Log
	unitOfWork      bool
	AccountID       string
	ApplicationID   string
	GroupID         string
	UserID          string
}

type DynamoEvent struct {
	ID            string `dynamo:"item_id,hash"`
	EntityID      string `dynamo:"entity_id"`
	EntityType    string `dynamo:"entity_type"`
	Payload       datatypes.JSON
	Type          string `dynamo:"type"`
	RootID        string `dynamo:"root_id"`
	ApplicationID string `dynamo:"application_id"`
	User          string `dynamo:"user"`
	SequenceNo    int64
}

var dynamo *dynamodb.DynamoDB

func init() {
	dynamo = connectDynamo()
}

// connectDynamo returns a dynamoDB client
func connectDynamo() (db *dynamodb.DynamoDB) {
	return dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
		Region:   aws.String("us-east-1"),
	})))
}

// CreateTable creates a table
func CreateTable() error {
	_, err := dynamo.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: aws.String("S"), // (S | N | B) for string, number, binary
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(TABLE_NAME),
	})

	return err
}

func NewDynamoEvent(event *Event) (DynamoEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return DynamoEvent{}, err
	}

	return DynamoEvent{
		ID:            event.ID,
		EntityID:      event.Meta.EntityID,
		EntityType:    event.Meta.EntityType,
		Payload:       payload,
		Type:          event.Type,
		RootID:        event.Meta.RootID,
		ApplicationID: event.Meta.Module,
		User:          event.Meta.User,
		SequenceNo:    event.Meta.SequenceNo,
	}, nil
}

//Persist - use with PutItem
func NewDynamoInput(event *Event) (*dynamodb.PutItemInput, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return &dynamodb.PutItemInput{}, err
	}

	return &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(event.ID),
			},
			"EntityID": {
				S: aws.String(event.Meta.EntityID),
			},
			"EntityType": {
				S: aws.String(event.Meta.EntityType),
			},
			"Payload": {
				B: payload,
			},
			"Type": {
				S: aws.String(event.Type),
			},
			"RootID": {
				S: aws.String(event.Meta.RootID),
			},
			"ApplicationID": {
				S: aws.String(event.Meta.Module),
			},
			"User": {
				S: aws.String(event.Meta.User),
			},
			"SequenceNo": {
				N: aws.String(strconv.Itoa(int(event.Meta.SequenceNo))),
			},
		},
		TableName: aws.String(TABLE_NAME),
	}, nil
}

//Persist - use with BatchWriteItem
func NewDynamoBatchInput(event *Event) (*dynamodb.WriteRequest, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return &dynamodb.WriteRequest{}, err
	}

	return &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: map[string]*dynamodb.AttributeValue{
				"ID": {
					S: aws.String(event.ID),
				},
				"EntityID": {
					S: aws.String(event.Meta.EntityID),
				},
				"EntityType": {
					S: aws.String(event.Meta.EntityType),
				},
				"Payload": {
					B: payload,
				},
				"Type": {
					S: aws.String(event.Type),
				},
				"RootID": {
					S: aws.String(event.Meta.RootID),
				},
				"ApplicationID": {
					S: aws.String(event.Meta.Module),
				},
				"User": {
					S: aws.String(event.Meta.User),
				},
				"SequenceNo": {
					N: aws.String(strconv.Itoa(int(event.Meta.SequenceNo))),
				},
			},
		},
	}, nil
}

func (d *EventRepositoryDynamo) Persist(ctxt context.Context, entity AggregateInterface) error {
	//Dynamo Batch cannot exceed 25 items
	count := 0
	total := 0

	var dynamoBatchEvents []*dynamodb.WriteRequest
	dynamoBatchWrite := &dynamodb.BatchWriteItemInput{}
	dynamoBatchWrite.RequestItems = make(map[string][]*dynamodb.WriteRequest)

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
			event.Meta.Module = d.ApplicationID
		}
		if event.Meta.Group == "" {
			event.Meta.Group = d.GroupID
		}

		dynamoEvent, err := NewDynamoBatchInput(event)
		if err != nil {
			return err
		}

		dynamoBatchEvents = append(dynamoBatchEvents, dynamoEvent)
		total++
		count++

		dynamoBatchWrite.RequestItems[TABLE_NAME] = dynamoBatchEvents

		if count == 25 {
			_, err := d.DB.BatchWriteItem(dynamoBatchWrite)
			if err != nil {
				return err
			}

			count = 0
			dynamoBatchEvents = []*dynamodb.WriteRequest{}
			dynamoBatchWrite.RequestItems = make(map[string][]*dynamodb.WriteRequest)

		}

	}

	if count != 0 && count < 25 {
		_, err := d.DB.BatchWriteItem(dynamoBatchWrite)
		if err != nil {
			return err
		}
	}

	//call persist on the aggregate root to clear the new changes array
	entity.Persist()

	for _, entity := range entities {
		d.eventDispatcher.Dispatch(ctxt, *entity.(*Event))
	}

	return nil
}

//AddSubscriber Allows you to add a handler that is triggered when events are dispatched
func (d *EventRepositoryDynamo) AddSubscriber(handler EventHandler) {
	d.eventDispatcher.AddSubscriber(handler)
}

//GetSubscribers Get the current list of event subscribers
func (d *EventRepositoryDynamo) GetSubscribers() ([]EventHandler, error) {
	return d.eventDispatcher.GetSubscribers(), nil
}

func (d *EventRepositoryDynamo) Flush() error {
	return nil
}

func (d *EventRepositoryDynamo) Migrate(ctx context.Context) error {
	return nil
}

//GetAggregateSequenceNumber gets the latest sequence number for the aggregate entity
func (d *EventRepositoryDynamo) GetAggregateSequenceNumber(ID string) (int64, error) {
	return 0, nil
}

//GetByAggregate get events for a root aggregate
func (d *EventRepositoryDynamo) GetByAggregate(ID string) ([]*Event, error) {
	tableName := TABLE_NAME

	input := &dynamodb.QueryInput{
		TableName: &tableName,
		IndexName: aws.String("ID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":RootID": {
				S: aws.String(ID),
			},
		},
		KeyConditionExpression: aws.String("RootID = :RootID"),
	}

	output, err := d.DB.Query(input)
	if err != nil {
		return nil, err
	}

	//Just to check if it is outputting something from the query
	if output == nil {
		return nil, err
	}

	var dEvents []*Event

	return dEvents, nil
}

func (d *EventRepositoryDynamo) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*Event, error) {
	return nil, nil
}

func (d *EventRepositoryDynamo) GetByEntityAndAggregate(EntityID string, Type string, RootID string) ([]*Event, error) {
	return nil, nil
}

//GetByAggregateAndType returns events given the entity id and the entity type.
//Deprecated: 08/12/2021 This was in theory returning events by entity (not necessarily root aggregate). Upon introducing the RootID
//events should now be retrieved by root id,entity type and entity id. Use GetByEntityAndAggregate instead
func (d *EventRepositoryDynamo) GetByAggregateAndType(ID string, entityType string) ([]*Event, error) {
	return nil, nil
}

func NewBasicEventRepositoryDynamo(dynamoDB *dynamodb.DynamoDB, logger Log, useUnitOfWork bool, accountID string, applicationID string) (EventRepository, error) {
	if useUnitOfWork {
		transaction := connectDynamo()
		return &EventRepositoryDynamo{DB: transaction, logger: logger, unitOfWork: useUnitOfWork, AccountID: accountID, ApplicationID: applicationID}, nil
	}
	return &EventRepositoryDynamo{DB: dynamoDB, logger: logger, AccountID: accountID, ApplicationID: applicationID}, nil
}

/*
var dynamo *dynamodb.DynamoDB

type TestEvent struct {
	ID     string `dynamo:"event_id,hash"`
	Name   string `dynamo:"name"`
	Random string `dynamo:"Random"`
}

func init() {
	dynamo = connectDynamo()
}

// connectDynamo returns a dynamoDB client
func connectDynamo() (db *dynamodb.DynamoDB) {
	return dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
		Region:   aws.String("us-east-1"),
	})))
}

// CreateTable creates a table
func CreateTable() error {
	_, err := dynamo.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: aws.String("S"), // (S | N | B) for string, number, binary
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("Events"),
	})

	return err
}

// PutEvent inserts the struct TestEvent
func PutEvent(event TestEvent) error {
	_, err := dynamo.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(event.ID),
			},
			"Name": {
				S: aws.String(event.Name),
			},
			"Random": {
				S: aws.String(event.Random),
			},
		},
		TableName: aws.String("Events"),
	})

	return err
}

func GetEvent(id string) (event TestEvent, err error) {
	result, err := dynamo.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
		TableName: aws.String("Events"),
	})

	if err != nil {
		return event, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &event)
	if err != nil {
		return event, err
	}

	return event, err

}
*/
