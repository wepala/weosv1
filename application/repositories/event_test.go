package repositories

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/wepala/weos/domain"
	//"gorm.io/gorm"
	"log"
	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "10.7", []string{"POSTGRES_USER=root", "POSTGRES_PASSWORD=secret", "POSTGRES_DB=test"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf("host=localhost port=%s user=root password=secret sslmode=disable database=test", resource.GetPort("5432/tcp")))
		//db, err = pgx.Connect(context.Background(),fmt.Sprintf("host=localhost port=%s user=root password=secret sslmode=disable database=test", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	//init.db
	//query, err := ioutil.ReadFile("./testdata/sql/init.sql")
	//if err != nil {
	//	panic(err)
	//}
	//if _, err := db.Exec(string(query)); err != nil {
	//	panic(err)
	//}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestEventRepositoryGorm_Persist(t *testing.T) {
	eventRepository, err := NewEventRepositoryWithGORM(db, nil)
	if err != nil {
		t.Fatalf("error encountered creating event repository '%s'", err)
	}
	err = eventRepository.Migrate()
	if err != nil {
		t.Fatalf("error encountered migration event repository '%s'", err)
	}

	mockEvent := &domain.Event{
		Type:    "TEST_EVENT",
		Payload: nil,
		Meta: domain.EventMeta{
			ID:          "some id",
			Application: "applicationID",
			Account:     "accountID",
			SequenceNo:  0,
		},
		Version: 1,
	}

	//add an event handler
	eventHandlerCalled := 0
	eventRepository.AddSubscriber(func(event domain.Event) {
		eventHandlerCalled += 1
	})

	err = eventRepository.Persist([]domain.Event{*mockEvent})
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	if eventHandlerCalled == 1 {
		t.Errorf("expected event handlers to be called %d time, called %d times", 1, eventHandlerCalled)
	}

	rows, err := db.Query("SELECT entity_id,type, account_id,application_id FROM gorm_events WHERE entity_id  = $1", "some id")
	if err != nil {
		t.Fatalf("error retrieving events '%s'", err)
	}

	for rows.Next() {
		var eventType, entityID, accountID, applicationID string
		err = rows.Scan(&entityID, &eventType, &accountID, &applicationID)
		if err != nil {
			t.Fatalf("error retrieving event '%s'", err)
		}

		if eventType != mockEvent.Type {
			t.Errorf("expected the type to be '%s', got '%s'", mockEvent.Type, eventType)
		}
	}
}
