package weos_test

import (
	"github.com/wepala/weos"
	"golang.org/x/net/context"
	"testing"
)

func TestGetAccount(t *testing.T) {
	t.Run("get account", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weos.ACCOUNT_ID, "123")
		accountID := weos.GetAccount(ctxt)
		if accountID != "123" {
			t.Errorf("expected account id to be '%s', got '%s'", "123", accountID)
		}
	})

	t.Run("get user", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weos.USER_ID, "123")
		user := weos.GetUser(ctxt)
		if user != "123" {
			t.Errorf("expected user id to be '%s', got '%s'", "123", user)
		}
	})

	t.Run("get log level", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weos.LOG_LEVEL, "123")
		logLevel := weos.GetLogLevel(ctxt)
		if logLevel != "123" {
			t.Errorf("expected log level to be '%s', got '%s'", "123", logLevel)
		}
	})

}
