package module_test

import (
	"context"
	"database/sql"
	_ "github.com/proullon/ramsql/driver"
	"github.com/wepala/weos/domain"
	"github.com/wepala/weos/module"
	"github.com/wepala/weos/persistence"
	"testing"
)

func TestNewApplicationFromConfig(t *testing.T) {
	config := &module.WeOSModuleConfig{
		ModuleID:  "1iPwGftUqaP4rkWdvFp6BBW2tOf",
		Title:     "Test Module",
		AccountID: "1iPwIGTgWVGyl4XfgrhCqYiiQ7d",
		Database: &module.WeOSDBConfig{
			Host:     "localhost",
			User:     "root",
			Password: "password",
			Port:     5432,
			Database: "test",
		},
		Log: &module.WeOSLogConfig{
			Level:        "debug",
			ReportCaller: false,
			Formatter:    "text",
		},
	}

	t.Run("basic module from config", func(t *testing.T) {
		app, err := module.NewApplicationFromConfig(config, nil, nil)
		if err != nil {
			t.Fatalf("error encountered setting up app")
		}
		if app.ModuleID != config.ModuleID {
			t.Errorf("expected the module id to be '%s', got '%s'", config.ModuleID, app.ModuleID)
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
		app, err := module.NewApplicationFromConfig(config, logger, nil)
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

		app, err := module.NewApplicationFromConfig(config, nil, db)
		if err != nil {
			t.Fatalf("error encountered setting up app")
		}
		if app.GetDBConnection().Ping() != nil {
			t.Errorf("didn't expect errors pinging the database")
		}
	})
}

func TestWeOSApp_AddProjection(t *testing.T) {
	config := &module.WeOSModuleConfig{
		ModuleID:  "1iPwGftUqaP4rkWdvFp6BBW2tOf",
		Title:     "Test Module",
		AccountID: "1iPwIGTgWVGyl4XfgrhCqYiiQ7d",
		Database: &module.WeOSDBConfig{
			Host:     "localhost",
			User:     "root",
			Password: "password",
			Port:     5432,
			Database: "test",
		},
		Log: &module.WeOSLogConfig{
			Level:        "debug",
			ReportCaller: false,
			Formatter:    "text",
		},
	}
	mockProjection := &ProjectionMock{
		GetEventHandlerFunc: func() persistence.EventHandler {
			return func(event domain.Event) {

			}
		},
		MigrateFunc: func(ctx context.Context) error {
			return nil
		},
	}
	app, err := module.NewApplicationFromConfig(config, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error occured setting up module '%s'", err)
	}

	err = app.AddProjection(mockProjection)
	if err != nil {
		t.Fatalf("unexpected error occured setting up projection '%s'", err)
	}

	err = app.Migrate(context.TODO())
	if err != nil {
		t.Fatalf("unexpected error running migrations '%s'", err)
	}

	if len(mockProjection.MigrateCalls()) != 1 {
		t.Errorf("expected the migrate function to be called %d time, called %d times", 1, len(mockProjection.MigrateCalls()))
	}

	//TODO confirm that the handler from the projection is added to the event repository IF one is configured
}