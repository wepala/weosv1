package weosv1_test

import (
	"github.com/wepala/weos"
	"golang.org/x/net/context"
	"testing"
)

func TestEventDisptacher_Dispatch(t *testing.T) {
	mockEvent := &weosv1.Event{
		Type:    "TEST_EVENT",
		Payload: nil,
		Meta: weosv1.EventMeta{
			EntityID: "some id",
			Module:   "applicationID",
			RootID:   "accountID",
		},
		Version: 1,
	}
	dispatcher := &weosv1.EventDisptacher{}
	handlersCalled := 0
	dispatcher.AddSubscriber(func(ctx context.Context, event weosv1.Event) {
		handlersCalled += 1
	})

	dispatcher.AddSubscriber(func(ctx context.Context, event weosv1.Event) {
		handlersCalled += 1
		if event.Type != mockEvent.Type {
			t.Errorf("expected the type to be '%s', got '%s'", mockEvent.Type, event.Type)
		}
	})
	dispatcher.Dispatch(context.TODO(), *mockEvent)

	if handlersCalled != 2 {
		t.Errorf("expected %d handler to be called, %d called", 2, handlersCalled)
	}
}
