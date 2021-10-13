package weos_test

import (
	"encoding/json"
	"testing"

	"github.com/wepala/weos"
)

type user struct {
	weos.AmorphousEntity
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
		if property.(*weos.StringProperty).UI != weos.UI_SINGLE_LINE {
			t.Errorf("expected UI to be '%s', got '%s'", weos.UI_SINGLE_LINE, property.(*weos.StringProperty).UI)
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
func TestAmorphousEntity_BooleanProperty(t *testing.T) {
	admin := new(user)
	t.Run("add property", func(t *testing.T) {
		admin.Set(new(weos.BooleanProperty).FromLabelAndValue("is trinidadian", true, false))
		property := admin.Get("is trinidadian")
		if property.GetType() != "boolean" {
			t.Errorf("expected type to be '%s', got '%s'", "boolean", property.GetType())
		}
		if !property.(*weos.BooleanProperty).Value {
			t.Errorf("expected '%s' to be true", property.GetLabel())
		}
		if property.(*weos.BooleanProperty).UI != weos.UI_CHECKBOX {
			t.Errorf("expected UI to be '%s', got '%s'", weos.UI_CHECKBOX, property.(*weos.BooleanProperty).UI)
		}
	})
}
func TestAmorphousEntity_NumericProperty(t *testing.T) {
	admin := new(user)
	t.Run("add property", func(t *testing.T) {
		admin.Set(new(weos.NumericProperty).FromLabelAndValue("amount", 100, false))
		property := admin.Get("amount")
		if property.GetType() != "numeric" {
			t.Errorf("expected type to be '%s', got '%s'", "numeric", property.GetType())
		}
		if property.(*weos.NumericProperty).Value != 100 {
			t.Errorf("expected value to be '%d', got '%f'", 100, property.(*weos.NumericProperty).Value)
		}
		if property.(*weos.NumericProperty).UI != weos.UI_SINGLE_LINE {
			t.Errorf("expected UI to be '%s', got '%s'", weos.UI_SINGLE_LINE, property.(*weos.NumericProperty).UI)
		}
	})
	t.Run("invalid numeric property", func(t *testing.T) {
		admin.Set(new(weos.NumericProperty).FromLabelAndValue("amount", 0, true))
		property := admin.Get("amount")
		if property.IsValid() {
			t.Fatalf("expected '%s' property to be invalid", property.GetLabel())
		}
	})
}

func TestAmorphousEntity_DeserializeJSON(t *testing.T) {
	admin := new(user)

	t.Run("test deserialize", func(t *testing.T) {
		admin.Set(new(weos.StringProperty).FromLabelAndValue("FirstName", "Eric", false))
		admin.Set(new(weos.BooleanProperty).FromLabelAndValue("isTrue", true, false))
		admin.Set(new(weos.NumericProperty).FromLabelAndValue("amount", 200, false))

		temp := &admin.AmorphousEntity
		marshall, err := json.Marshal(temp)
		if err != nil {
			t.Errorf("Unexpected err marshalling amorphous entity, '%s'", err)
		}

		var someUser user
		json.Unmarshal(marshall, &someUser)

		if someUser.Get("FirstName").(*weos.StringProperty).Value != "Eric" {
			t.Errorf("some user was not unmarshalled correctly")
		}
	})
}
