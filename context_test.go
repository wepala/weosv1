package weosv1_test

import (
	"github.com/wepala/weosv1"
	"golang.org/x/net/context"
	"testing"
)

func TestGetAccount(t *testing.T) {
	t.Run("get account", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weosv1.ACCOUNT_ID, "123")
		accountID := weosv1.GetAccount(ctxt)
		if accountID != "123" {
			t.Errorf("expected account id to be '%s', got '%s'", "123", accountID)
		}
	})

	t.Run("get user", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weosv1.USER_ID, "123")
		user := weosv1.GetUser(ctxt)
		if user != "123" {
			t.Errorf("expected user id to be '%s', got '%s'", "123", user)
		}
	})

	t.Run("get log level", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weosv1.LOG_LEVEL, "123")
		logLevel := weosv1.GetLogLevel(ctxt)
		if logLevel != "123" {
			t.Errorf("expected log level to be '%s', got '%s'", "123", logLevel)
		}
	})

	t.Run("get request id", func(t *testing.T) {
		ctxt := context.WithValue(context.Background(), weosv1.REQUEST_ID, "123")
		requestID := weosv1.GetRequestID(ctxt)
		if requestID != "123" {
			t.Errorf("expected request id to be '%s', got '%s'", "123", requestID)
		}
	})

}
