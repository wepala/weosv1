package weos_test

import (
	"github.com/wepala/weos"
	"testing"
)

func TestEventDisptacher_Dispatch(t *testing.T) {
	mockEvent := &weos.Event{
		Type:    "TEST_EVENT",
		Payload: nil,
		Meta: weos.EventMeta{
			EntityID: "some id",
			Module:   "applicationID",
			Account:  "accountID",
		},
		Version: 1,
	}
	dispatcher := &weos.EventDisptacher{}
	handlersCalled := 0
	dispatcher.AddSubscriber(func(event weos.Event) {
		handlersCalled += 1
	})

	dispatcher.AddSubscriber(func(event weos.Event) {
		handlersCalled += 1
		if event.Type != mockEvent.Type {
			t.Errorf("expected the type to be '%s', got '%s'", mockEvent.Type, event.Type)
		}
	})
	dispatcher.Dispatch(*mockEvent)

	if handlersCalled != 2 {
		t.Errorf("expected %d handler to be called, %d called", 2, handlersCalled)
	}
}