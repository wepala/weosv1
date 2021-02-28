package module

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
	"time"
)

//go:generate moq -out aws_mocks_test.go -pkg module_test /Users/akeem/go/pkg/mod/github.com/aws/aws-sdk-go@v1.37.20/service/sqs/sqsiface SQSAPI

type SQSService struct {
	Client   sqsiface.SQSAPI
	queueURL string
}

type SQSReceiverCallback func(output *sqs.ReceiveMessageOutput) *Command

func (s *SQSService) Dispatch(ctx context.Context, command *Command) error {
	id := ksuid.New().String()
	payloadBytes, err := json.Marshal(&command.Payload)
	if err != nil {
		return err
	}
	payload := string(payloadBytes)

	_, err = s.Client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		DelaySeconds: aws.Int64(int64(command.Metadata.ExecutionDate.Sub(time.Now()).Seconds())),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(command.Type),
			},
			"accountId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(command.Metadata.AccountID),
			},
			"applicationId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(command.Metadata.ApplicationID),
			},
			"userId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(command.Metadata.UserID),
			},
		},
		MessageBody:            &payload,
		MessageDeduplicationId: &id,
		QueueUrl:               &s.queueURL,
	})

	return err
}

func (s *SQSService) Receive(ctx context.Context, callback SQSReceiverCallback) (*Command, error) {
	output, err := s.Client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		AttributeNames:        nil,
		MessageAttributeNames: nil,
		QueueUrl:              nil,
		VisibilityTimeout:     nil,
		WaitTimeSeconds:       nil,
	})
	if err != nil {
		return nil, err
	}

	return callback(output), nil
}

func NewSQSService(queueURL string) *SQSService {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)
	return &SQSService{
		Client:   svc,
		queueURL: queueURL,
	}
}
