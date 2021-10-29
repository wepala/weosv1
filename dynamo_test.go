package weos_test

import (
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/wepala/weos"
	"golang.org/x/net/context"
)

var dynamoDB *dynamodb.DynamoDB
var database = flag.String("database", "dynamoDB", "run dynamo tests")

func TestMain(m *testing.M) {
	flag.Parse()
	switch *database {
	case "dynamoDB":
		log.Infof("Started DynamoDB")
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			Name:         "dynamodb-local",
			ExposedPorts: []string{"8000:8000/tcp"},
			Env:          map[string]string{"AWS_ACCESS_KEY_ID": "fakeMyKeyId", "AWS_SECRET_ACCESS_KEY": "fakeSecretAccessKey"},
			//WaitingFor:   wait.ForLog("Started DynamoDB"),
			//Cmd:          []string{"java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb -dbPath ./data"},

		}
		esContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start dynamo db container '%s'", err)
		}

		defer esContainer.Terminate(ctx)

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
			/*Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
				AccessKeyID:     *aws.String("fakeMyKeyId"),
				SecretAccessKey: *aws.String("fakeSecretAccessKey"),
			}),*/
			Endpoint: aws.String("http://localhost:8000"),
		})
		if err != nil {
			log.Fatalf("Could not setup dynamoDB: %s", err.Error())
		}
		dynamoDB = dynamodb.New(sess)

		code := m.Run()

		//os.Remove("shared-local-instance.db")
		os.Exit(code)
	}
}

func TestDynamo_AddEvent(t *testing.T) {

	eventRepository, err := weos.NewBasicEventRepositoryDynamo(dynamoDB, log.New(), false, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	err = eventRepository.(*weos.EventRepositoryDynamo).Migrate(context.Background())
	if err != nil {
		t.Fatalf("error setting up application'%s'", err)
	}

	generateEvents := make([]*weos.Event, 50000)
	entity := &weos.AggregateRoot{}

	for i := 0; i < 50000; i++ {

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
		t.Fatalf("error persisting events '%s'", err)
	}

	//Check for events
}

func TestDynamo_GetByEntityAndAggregate(t *testing.T) {
	eventRepository, err := weos.NewBasicEventRepositoryDynamo(dynamoDB, log.New(), false, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	err = eventRepository.(*weos.EventRepositoryDynamo).Migrate(context.Background())
	if err != nil {
		t.Fatalf("error setting up application'%s'", err)
	}

	generateEvents := make([]*weos.Event, 5)
	entity := &weos.AggregateRoot{}

	for i := 0; i < 5; i++ {

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
		t.Fatalf("error persisting events '%s'", err)
	}

	output, err := eventRepository.GetByEntityAndAggregate("batch id", "SomeType", "12345")
	if err != nil {
		t.Fatalf("error getting by aggregate '%s'", err)
	}

	if output == nil {
		t.Fatalf("expected output to not be nil")
	}

	if output[0].ID != generateEvents[0].ID {
		t.Fatalf("error getting by aggregate: expected '%s' but got '%s'", generateEvents[0].ID, output[0].ID)
	}

	if output[0].Type != generateEvents[0].Type {
		t.Fatalf("error getting by aggregate: expected '%s' but got '%s'", generateEvents[0].Type, output[0].Type)
	}

	if output[3].ID != generateEvents[3].ID {
		t.Fatalf("error getting by aggregate: expected '%s' but got '%s'", generateEvents[3].ID, output[3].ID)
	}

	if output[3].Type != generateEvents[3].Type {
		t.Fatalf("error getting by aggregate: expected '%s' but got '%s'", generateEvents[3].Type, output[3].Type)
	}

}
