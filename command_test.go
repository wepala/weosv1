package weosv1_test

import (
	"github.com/wepala/weosv1"
	"golang.org/x/net/context"
	"testing"
)

func TestCommandDisptacher_Dispatch(t *testing.T) {
	mockCommand := &weosv1.Command{
		Type:    "TEST_COMMAND",
		Payload: nil,
		Metadata: weosv1.CommandMetadata{
			Version: 1,
		},
	}
	dispatcher := &weosv1.DefaultCommandDispatcher{}
	handlersCalled := 0
	dispatcher.AddSubscriber(mockCommand, func(ctx context.Context, command *weosv1.Command) error {
		handlersCalled += 1
		return nil
	})

	dispatcher.AddSubscriber(&weosv1.Command{
		Type:     "*",
		Payload:  nil,
		Metadata: weosv1.CommandMetadata{},
	}, func(context context.Context, event *weosv1.Command) error {
		handlersCalled += 1
		if event.Type != mockCommand.Type {
			t.Errorf("expected the type to be '%s', got '%s'", mockCommand.Type, event.Type)
		}
		return nil
	})
	dispatcher.Dispatch(context.TODO(), mockCommand)

	if handlersCalled != 2 {
		t.Errorf("expected %d handler to be called, %d called", 2, handlersCalled)
	}
}
