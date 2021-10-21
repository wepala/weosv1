package weos

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var dynamo *dynamodb.DynamoDB

type TestEvent struct {
	ID   string `dynamo:"event_id,hash"`
	Name string `dynamo:"name"`
}

func init() {
	dynamo = connectDynamo()
}

// connectDynamo returns a dynamoDB client
func connectDynamo() (db *dynamodb.DynamoDB) {
	return dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))
}

// CreateTable creates a table
func CreateTable() error {
	_, err := dynamo.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       aws.String("HASH"),
			},
		},
		TableName: aws.String("Events"),
	})

	return err
}

// PutItem inserts the struct Person
func PutItem(event TestEvent) error {
	_, err := dynamo.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Id": {
				N: aws.String(event.ID),
			},
			"Name": {
				S: aws.String(event.Name),
			},
		},
		TableName: aws.String("Events"),
	})

	return err
}

/*type TestEvent struct {
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
}*/
