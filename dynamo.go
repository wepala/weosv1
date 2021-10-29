package weos

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
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

const TABLE_NAME = "Events"

type EventRepositoryDynamo struct {
	DB              *dynamodb.DynamoDB
	eventDispatcher EventDisptacher
	logger          Log
	AccountID       string
	ApplicationID   string
	GroupID         string
	UserID          string
}

type DynamoEvent struct {
	EventID       string `dynamo:"event_id"`
	EntityID      string `dynamo:"entity_id"`
	EntityType    string `dynamo:"entity_type"`
	Payload       datatypes.JSON
	Type          string `dynamo:"type"`
	RootID        string `dynamo:"root_id"`
	ApplicationID string `dynamo:"application_id"`
	User          string `dynamo:"user"`
	SequenceNo    int64
}

// connectDynamo returns a dynamoDB client
/*func connectDynamo() (db *dynamodb.DynamoDB) {
	return dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
		Region:   aws.String("us-east-1"),
	})))
}*/

func NewDynamoEvent(event *Event) (DynamoEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return DynamoEvent{}, err
	}

	return DynamoEvent{
		EventID:       event.ID,
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
			"EventID": {
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
				"EventID": {
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
				"PartitionKey": {
					S: aws.String(event.Meta.RootID + "-" + event.Meta.EntityType + "-" + event.Meta.EntityID),
				},
				"SortKey": {
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

		if count == 25 {
			dynamoBatchWrite.RequestItems[TABLE_NAME] = dynamoBatchEvents
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
		dynamoBatchWrite.RequestItems[TABLE_NAME] = dynamoBatchEvents
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

//TODO - This needs work. (I just copied and pasted the create table func.)
func (d *EventRepositoryDynamo) Migrate(ctx context.Context) error {
	existingTables, err := d.DB.ListTables(
		&dynamodb.ListTablesInput{},
	)
	if err != nil {
		return err
	}
	//Checks if the table exists before attempting to create it
	if len(existingTables.TableNames) != 0 {
		return nil
	}

	_, err = d.DB.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(TABLE_NAME),
		// PK - ConcatKey + Seq No
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				//Rename to PartitionKey - ConcatKey
				AttributeName: aws.String("PartitionKey"),
				AttributeType: aws.String("S"), // (S | N | B) for string, number, binary
			},
			{
				AttributeName: aws.String("SortKey"),
				AttributeType: aws.String("N"), // (S | N | B) for string, number, binary
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PartitionKey"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("SortKey"),
				KeyType:       aws.String("Range"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})

	return err
}

//GetAggregateSequenceNumber gets the latest sequence number for the aggregate entity
func (d *EventRepositoryDynamo) GetAggregateSequenceNumber(ID string) (int64, error) {
	return 0, nil
}

//GetByAggregate get events for a root aggregate
func (d *EventRepositoryDynamo) GetByAggregate(ID string) ([]*Event, error) {
	return nil, nil
}

func (d *EventRepositoryDynamo) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*Event, error) {
	return nil, nil
}

func (d *EventRepositoryDynamo) GetByEntityAndAggregate(EntityID string, Type string, RootID string) ([]*Event, error) {

	input := &dynamodb.QueryInput{
		TableName: aws.String(TABLE_NAME),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":PartitionKey": {
				S: aws.String(RootID + "-" + Type + "-" + EntityID),
			},
		},
		KeyConditionExpression: aws.String("PartitionKey = :PartitionKey"),
	}

	output, err := d.DB.Query(input)
	if err != nil {
		return nil, err
	}

	//Just to check if it is outputting something from the query
	if output == nil {
		return nil, err
	}

	var dynamoEvents []*DynamoEvent
	var events []*Event

	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &dynamoEvents)
	if err != nil {
		return nil, err
	}

	//Converts dynamoEvent to regular Event
	for _, event := range dynamoEvents {
		events = append(events, &Event{
			ID:      event.EventID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta: EventMeta{
				EntityID:   event.EntityID,
				EntityType: event.EntityType,
				RootID:     event.RootID,
				Module:     event.ApplicationID,
				User:       event.User,
				SequenceNo: event.SequenceNo,
			},
			Version: 0,
		})
	}

	return events, nil
}

//GetByAggregateAndType returns events given the entity id and the entity type.
//Deprecated: 08/12/2021 This was in theory returning events by entity (not necessarily root aggregate). Upon introducing the RootID
//events should now be retrieved by root id,entity type and entity id. Use GetByEntityAndAggregate instead
func (d *EventRepositoryDynamo) GetByAggregateAndType(ID string, entityType string) ([]*Event, error) {
	return nil, nil
}

func NewBasicEventRepositoryDynamo(dynamoDB *dynamodb.DynamoDB, logger Log, accountID string, applicationID string) (EventRepository, error) {
	return &EventRepositoryDynamo{DB: dynamoDB, logger: logger, AccountID: accountID, ApplicationID: applicationID}, nil
}
