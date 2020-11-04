package domain_test

import (
	"testing"

	"github.com/wepala/weos/domain"
)

func TestNewBasicEvent(t *testing.T) {
	event, _ := domain.NewBasicEvent("TEST_EVENT", "1iNqlx5htN0oJ3viyfWkAofJX7k", nil, 0)
	if event.Type != "TEST_EVENT" {
		t.Errorf("expected event to be type '%s', got '%s'", "TEST_EVENT", event.Type)
	}

	if event.Meta.EntityID != "1iNqlx5htN0oJ3viyfWkAofJX7k" {
		t.Errorf("expected the entity id to be '%s', got'%s'", "1iNqlx5htN0oJ3viyfWkAofJX7k", event.Meta.EntityID)
	}
}

func TestEvent_IsValid(t *testing.T) {
	t.Run("valid event", func(t *testing.T) {
		event, _ := domain.NewBasicEvent("TEST_EVENT", "1iNqlx5htN0oJ3viyfWkAofJX7k", nil, 0)
		if !event.IsValid() {
			t.Errorf("expected the event to be valid")
		}
	})

	t.Run("no entity id, event invalid", func(t *testing.T) {
		event, _ := domain.NewBasicEvent("TEST_EVENT", "", nil, 0)
		if event.IsValid() {
			t.Fatalf("expected the event to be invalid")
		}

		if len(event.GetErrors()) == 0 {
			t.Errorf("expected the event to have errors")
		}
	})

	t.Run("no event id, event invalid", func(t *testing.T) {
		event := domain.Event{
			ID:      "",
			Type:    "Some Type",
			Payload: nil,
			Meta: domain.EventMeta{
				EntityID: "1iNqlx5htN0oJ3viyfWkAofJX7k",
			},
			Version: 1,
		}
		if event.IsValid() {
			t.Fatalf("expected the event to be invalid")
		}

		if len(event.GetErrors()) == 0 {
			t.Errorf("expected the event to have errors")
		}
	})

	t.Run("no version no, event invalid", func(t *testing.T) {
		event := domain.Event{
			ID:      "adfasdf",
			Type:    "Some Type",
			Payload: nil,
			Meta: domain.EventMeta{
				EntityID: "1iNqlx5htN0oJ3viyfWkAofJX7k",
			},
			Version: 0,
		}
		if event.IsValid() {
			t.Fatalf("expected the event to be invalid")
		}

		if len(event.GetErrors()) == 0 {
			t.Errorf("expected the event to have errors")
		}
	})

	t.Run("no type, event invalid", func(t *testing.T) {
		event := domain.Event{
			ID:      "adfasdf",
			Type:    "",
			Payload: nil,
			Meta: domain.EventMeta{
				EntityID: "1iNqlx5htN0oJ3viyfWkAofJX7k",
			},
			Version: 1,
		}
		if event.IsValid() {
			t.Fatalf("expected the event to be invalid")
		}

		if len(event.GetErrors()) == 0 {
			t.Errorf("expected the event to have errors")
		}
	})

}
