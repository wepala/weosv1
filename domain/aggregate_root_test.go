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

	mockEvent := &domain.Event{
		Type: "Event",
		Payload: &struct {
			Title string
		}{Title: "Test"},
		Meta: domain.EventMeta{
			EntityID: "",
		},
		Version: 0,
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = domain.DefaultReducer(baseAggregate, *mockEvent, nil).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}
}

func TestAggregateRoot_NewAggregateFromEvents(t *testing.T) {
	type BaseAggregate struct {
		domain.AggregateRoot
		Title string `json:"title"`
	}

	mockEvent := &domain.Event{
		Type: "Event",
		Payload: &struct {
			Title string
		}{Title: "Test"},
		Meta: domain.EventMeta{
			EntityID: "",
		},
		Version: 1,
	}
	baseAggregate := &BaseAggregate{}
	baseAggregate = domain.NewAggregateFromEvents(baseAggregate, []domain.Event{*mockEvent}).(*BaseAggregate)
	if baseAggregate.Title != "Test" {
		t.Errorf("expected aggregate title to be '%s', got '%s'", "Test", baseAggregate.Title)
	}
}
