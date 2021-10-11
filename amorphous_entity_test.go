package weos_test

import (
	"github.com/wepala/weos"
	"testing"
)

type user struct {
	*weos.AmorphousEntity
}

func TestAmorphousEntity_StringProperty(t *testing.T) {
	admin := new(user)
	t.Run("add property", func(t *testing.T) {
		admin.Set(new(weos.StringProperty).FromLabelAndValue("FirstName", "Eric", false))
		property := admin.Get("FirstName")
		if property.GetType() != "string" {
			t.Errorf("expected type to be '%s', got '%s'", "string", property.GetType())
		}
		if property.(*weos.StringProperty).Value != "Eric" {
			t.Errorf("expected value to be '%s', got '%s'", "Eric", property.(*weos.StringProperty).Value)
		}
	})
	t.Run("invalid string property", func(t *testing.T) {
		admin.Set(new(weos.StringProperty).FromLabelAndValue("FirstName", "", true))
		property := admin.Get("FirstName")
		if property.IsValid() {
			t.Fatalf("expected '%s' property to be invalid", property.GetLabel())
		}
	})
}
