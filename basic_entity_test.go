package weos_test

import (
	"errors"
	"github.com/wepala/weos"
	"testing"
)

func TestBasicEntity_AddError(t *testing.T) {
	entity := &weos.BasicEntity{}
	entity.AddError(errors.New("some error"))
	if len(entity.GetErrors()) != 1 {
		t.Errorf("expected the length of error to be %d, got %d", 1, len(entity.GetErrors()))
	}
}
