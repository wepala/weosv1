package repositories_test

import (
	"github.com/wepala/weos/domain"
	"github.com/wepala/weos/infrastructure/repositories"
	"testing"
)

func TestEventDisptacher_Dispatch(t *testing.T) {
	mockEvent := &domain.Event{
		Type:    "TEST_EVENT",
		Payload: nil,
		Meta: domain.EventMeta{
			EntityID:    "some id",
			Application: "applicationID",
			Account:     "accountID",
		},
		Version: 1,
	}
	dispatcher := &repositories.EventDisptacher{}
	handlersCalled := 0
	dispatcher.AddSubscriber(func(event domain.Event) {
		handlersCalled += 1
	})

	dispatcher.AddSubscriber(func(event domain.Event) {
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
