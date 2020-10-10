package domain_test

import (
	"errors"
	"github.com/wepala/weos/domain"
	"testing"
)

func TestBasicEntity_AddError(t *testing.T) {
	entity := &domain.BasicEntity{}
	entity.AddError(errors.New("some error"))
	if len(entity.GetErrors()) != 1 {
		t.Errorf("expected the length of error to be %d, got %d", 1, len(entity.GetErrors()))
	}
}
