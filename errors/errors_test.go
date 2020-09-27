package errors_test

import (
	"errors"
	errors2 "github.com/wepala/weos/errors"
	"testing"
)

func TestErrorFactory_NewDomainError(t *testing.T) {
	err := errors2.NewDomainError("some domain error", "User", "1", errors.New("some other error"))
	if err.Unwrap().Error() != "some other error" {
		t.Errorf("expected embedded error to be %s, got %s", "some other error", err.Unwrap().Error())
	}

	if err.Error() != "some domain error" {
		t.Errorf("expected the error to be %s, got %s", "some domain error", err.Error())
	}
}