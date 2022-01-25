package weosv1_test

import (
	"errors"
	"github.com/wepala/weos"
	"testing"
)

func TestBasicEntity_AddError(t *testing.T) {
	entity := &weosv1.BasicEntity{}
	entity.AddError(errors.New("some error"))
	if len(entity.GetErrors()) != 1 {
		t.Errorf("expected the length of error to be %d, got %d", 1, len(entity.GetErrors()))
	}
}
