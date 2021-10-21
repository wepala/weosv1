package weos_test

import (
	"testing"

	"github.com/wepala/weos"
)

func TestDynamo_CreateEvent(t *testing.T) {

	testEvent := &weos.TestEvent{
		Name: "Test Event 1",
	}

	service, err := weos.NewEventService("Events")
	if err != nil {
		t.Fatalf("error creating service '%s'", err)
	}

	err = service.CreateEvent(testEvent)
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}
}

func TestDynamo_GetEvent(t *testing.T) {

	testEvent := &weos.TestEvent{
		Name: "Test Event 1",
	}

	service, err := weos.NewEventService("Events")
	if err != nil {
		t.Fatalf("error creating service '%s'", err)
	}

	err = service.CreateEvent(testEvent)
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}

	err = service.GetEvent(testEvent)
	if err != nil {
		t.Fatalf("error getting event '%s'", err)
	}
}
