package weosv1_test

import (
	"github.com/wepala/weos"
	"testing"
)

func TestAggregateRoot_DefaultReducer(t *testing.T) {
	type BaseAggregate struct {
		weosv1.AggregateRoot
		Title string `json:"title"`
	}

	mockEvent, err := weosv1.NewBasicEvent("Event", "", "BaseAggregate", &struct {
		Title string `json:"title"`
	}{Title: "Test"})
	if err != nil {
		t.Fatalf("error creating mock event '%s'", err)
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = weosv1.DefaultReducer(baseAggregate, mockEvent, nil).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}
}

func TestAggregateRoot_NewAggregateFromEvents(t *testing.T) {
	type BaseAggregate struct {
		weosv1.AggregateRoot
		Title string `json:"title"`
	}

	mockEvent, err := weosv1.NewBasicEvent("Event", "", "BaseAggregate", &struct {
		Title string `json:"title"`
	}{Title: "Test"})
	if err != nil {
		t.Fatalf("error creating mock event '%s'", err)
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = weosv1.NewAggregateFromEvents(baseAggregate, []*weosv1.Event{mockEvent}).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}
}
