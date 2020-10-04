package application_test

import (
	"database/sql"
	_ "github.com/proullon/ramsql/driver"
	"github.com/wepala/weos/application"
	"testing"
)

func TestNewApplicationFromConfig(t *testing.T) {
	config := &application.WeOSApplicationConfig{
		ApplicationID:    "1iPwGftUqaP4rkWdvFp6BBW2tOf",
		ApplicationTitle: "Test Application",
		AccountID:        "1iPwIGTgWVGyl4XfgrhCqYiiQ7d",
		Database: &application.WeOSDBConfig{
			Host:     "localhost",
			User:     "root",
			Password: "password",
			Port:     5432,
			Database: "test",
		},
		Log: &application.WeOSLogConfig{
			Level:        "debug",
			ReportCaller: false,
			Formatter:    "text",
		},
	}

	t.Run("basic application from config", func(t *testing.T) {
		app, err := application.NewApplicationFromConfig(config, nil, nil)
		if err != nil {
			t.Fatalf("error encountered setting up app")
		}
		if app.ApplicationID != config.ApplicationID {
			t.Errorf("expected the application id to be '%s', got '%s'", config.ApplicationID, app.ApplicationID)
		}

		if app.GetDBConnection() == nil {
			t.Error("expected the db connection to be setup")
		}

		if app.GetLogger() == nil {
			t.Error("expected the logger to be setup")
		}
	})

	t.Run("override logger", func(t *testing.T) {
		logger := &Ext1FieldLoggerMock{
			DebugFunc: func(args ...interface{}) {

			},
		}
		app, err := application.NewApplicationFromConfig(config, logger, nil)
		if err != nil {
			t.Fatalf("error encountered setting up app")
		}
		app.GetLogger().Debug("some debug")
		if len(logger.DebugCalls()) == 0 {
			t.Errorf("expected the debug function on the logger to be called at least %d time, called %d times", 1, len(logger.DebugCalls()))
		}
	})

	t.Run("override db connection", func(t *testing.T) {
		db, err := sql.Open("ramsql", "TestLoadUserAddresses")
		if err != nil {
			t.Fatalf("sql.Open : Error : %s\n", err)
		}
		defer db.Close()

		app, err := application.NewApplicationFromConfig(config, nil, db)
		if err != nil {
			t.Fatalf("error encountered setting up app")
		}
		if app.GetDBConnection().Ping() != nil {
			t.Errorf("didn't expect errors pinging the database")
		}
	})
}
