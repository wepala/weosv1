package weos

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

/*COMMANDS -
docker-compose up

For AWS CLI cmds, add --endpoint-url http://localhost:8000 at the end to hit local dynamo

aws dynamodb list-tables --endpoint-url http://localhost:8000
aws dynamodb scan --table-name Events --endpoint-url http://localhost:8000
aws dynamodb delete-table --table-name Events --endpoint-url http://localhost:8000
*/

/*type DynamoEvent struct {
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
