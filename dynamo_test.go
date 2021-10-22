package weos_test

import (
	"testing"

	"github.com/wepala/weos"
)

func TestDynamo_CreateTable(t *testing.T) {

	err := weos.CreateTable()
	if err != nil {
		t.Fatalf("error creating table '%s'", err)
	}
}

func TestDynamo_AddEvent(t *testing.T) {

	testEvent := weos.TestEvent{
		ID:     "3",
		Name:   "Test Event 3",
		Random: "12345",
	}

	err := weos.PutEvent(testEvent)
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}
}

func TestDynamo_GetEvent(t *testing.T) {

	testEvent, err := weos.GetEvent("2")
	if err != nil {
		t.Fatalf("error creating event '%s'", err)
	}

	if testEvent.ID != "2" {
		t.Fatalf("Expected test event ID to be '%s', got '%s'", "2", testEvent.ID)
	}
	if testEvent.Name != "Test Event 2" {
		t.Fatalf("Expected test event name to be '%s', got '%s'", "Test Event 2", testEvent.Name)
	}

}
