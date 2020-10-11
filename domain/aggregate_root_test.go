package domain_test

import (
	"github.com/wepala/weos/domain"
	"testing"
)

func TestAggregateRoot_DefaultReducer(t *testing.T) {
	type BaseAggregate struct {
		domain.AggregateRoot
		Title string `json:"title"`
	}

	mockEvent, err := domain.NewBasicEvent("Event", "", &struct {
		Title string `json:"title"`
	}{Title: "Test"}, "")
	if err != nil {
		t.Fatalf("error creating mock event '%s'", err)
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = domain.DefaultReducer(baseAggregate, mockEvent, nil).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}
}

func TestAggregateRoot_NewAggregateFromEvents(t *testing.T) {
	type BaseAggregate struct {
		domain.AggregateRoot
		Title string `json:"title"`
	}

	mockEvent, err := domain.NewBasicEvent("Event", "", &struct {
		Title string `json:"title"`
	}{Title: "Test"}, "123")
	if err != nil {
		t.Fatalf("error creating mock event '%s'", err)
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = domain.NewAggregateFromEvents(baseAggregate, []*domain.Event{mockEvent}).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}

	if baseAggregate.GetUser().ID != "123" {
		t.Errorf("expectected user id to be '%s', got '%s'", "123", baseAggregate.GetUser().ID)
	}
}
