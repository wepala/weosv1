package weos

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
		TableName: aws.String("Events"),
	}, nil
}

func (d *EventRepositoryDynamo) PersistDynamo(ctxt context.Context, entity AggregateInterface) error {
	//TODO use the information in the context to get account info, module info. //didn't think it should barf if an empty list is passed
	var dynamoEvents []DynamoEvent
	entities := entity.GetNewChanges()
	/*savePointID := "s" + ksuid.New().String() //NOTE the save point can't start with a number
	e.logger.Infof("persisting %d events with save point %s", len(entities), savePointID)
	if e.unitOfWork {
		e.DB.SavePoint(savePointID)
	}*/
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
		/*if !event.IsValid() {
			for _, terr := range event.GetErrors() {
				d.logger.Errorf("error encountered persisting entity '%s', '%s'", event.Meta.EntityID, terr)
			}
			if d.unitOfWork {
				d.logger.Debugf("rolling back saving events to %s", savePointID)
				d.DB.RollbackTo(savePointID)
			}

			return event.GetErrors()[0]
		}*/

		dynamoEvent, err := NewDynamoEvent(event)
		if err != nil {
			return err
		}
		dynamoEvents = append(dynamoEvents, dynamoEvent)
	}

	//Use NewDynamoInput and run the PutItem for each event? Research batchInputs

	//PutItem
	db := d.DB.PutItem(dynamoEvents)
	if db.Error != nil {
		return db.Error
	}

	//call persist on the aggregate root to clear the new changes array
	entity.Persist()

	for _, entity := range entities {
		e.eventDispatcher.Dispatch(ctxt, *entity.(*Event))
	}
	return nil
}

/*// PutEvent inserts the struct TestEvent
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
}*/

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
