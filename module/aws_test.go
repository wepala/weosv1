package module_test

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/wepala/weos/module"
	"testing"
	"time"
)

func TestSQSDispatcher_Dispatch(t *testing.T) {
	disptacher := module.NewSQSService("someurl")
	executionTime := time.Now().Add(10 * time.Minute)
	payload := &struct {
		Title string `json:"title"`
	}{
		Title: "Test",
	}
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}
	mockCommand := &module.Command{
		Type:    "weos.test",
		Payload: marshalledPayload,
		Execute: nil,
		Metadata: module.CommandMetadata{
			Version:       1,
			ApplicationID: "test-application",
			AccountID:     "account-id",
			UserID:        "user-id",
			ExecutionDate: &executionTime,
		},
	}
	mockSQSClient := SQSAPIMock{
		SendMessageWithContextFunc: func(contextMoqParam context.Context, sendMessageInput *sqs.SendMessageInput, options ...request.Option) (*sqs.SendMessageOutput, error) {
			//check that the message is structured the way we expect
			err := json.Unmarshal([]byte(*sendMessageInput.MessageBody), &payload)
			if err != nil {
				t.Fatalf("error encountered unmarshalling SQS message body '%s'", err)
			}

			if payload.Title != "Test" {
				t.Errorf("expected the message body to be '%s', got '%s'", "Test", payload.Title)
			}

			messageType, ok := sendMessageInput.MessageAttributes["type"]
			if !ok {
				t.Errorf("expected message to have an attribute 'type'")
			} else {
				if *messageType.StringValue != mockCommand.Type {
					t.Errorf("expected the type to be '%s', got '%s'", mockCommand.Type, *messageType.StringValue)
				}
			}

			applicationID, ok := sendMessageInput.MessageAttributes["applicationId"]
			if !ok {
				t.Errorf("expected message to have an attribute 'applicationId'")
			} else {
				if *applicationID.StringValue != mockCommand.Metadata.ApplicationID {
					t.Errorf("expected the application id to be '%s', got '%s'", mockCommand.Metadata.ApplicationID, *applicationID.StringValue)
				}
			}

			accountID, ok := sendMessageInput.MessageAttributes["accountId"]
			if !ok {
				t.Errorf("expected message to have an attribute 'accountId'")
			} else {
				if *accountID.StringValue != mockCommand.Metadata.AccountID {
					t.Errorf("expected the account id to be '%s', got '%s'", mockCommand.Metadata.AccountID, *accountID.StringValue)
				}
			}

			userID, ok := sendMessageInput.MessageAttributes["userId"]
			if !ok {
				t.Errorf("expected message to have an attribute 'userID'")
			} else {
				if *userID.StringValue != mockCommand.Metadata.UserID {
					t.Errorf("expected the user id to be '%s', got '%s'", mockCommand.Metadata.UserID, *userID.StringValue)
				}
			}

			if sendMessageInput.DelaySeconds == aws.Int64(0) {
				t.Errorf("expected there to be a delay on execution of %d, got %d", aws.Int64(600), sendMessageInput.DelaySeconds)
			}

			return nil, nil
		},
	}
	disptacher.Client = &mockSQSClient
	err = disptacher.Dispatch(context.TODO(), mockCommand)
	if err != nil {
		t.Fatalf("unexecpted error occured %s", err)
	}

	if len(mockSQSClient.SendMessageWithContextCalls()) != 1 {
		t.Errorf("expected send message with context to be called %d time, called %d times", 1, len(mockSQSClient.SendMessageWithContextCalls()))
	}
}
