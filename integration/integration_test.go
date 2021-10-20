package integration_test

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *sql.DB
var gormDB *gorm.DB
var database = flag.String("database", "sqlite3", "run database integration tests")
var err error

func TestMain(m *testing.M) {
	flag.Parse()
	switch *database {
	case "postgres":
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err.Error())
		}

		// pulls an image, creates a container based on it and runs it
		resource, err := pool.Run("postgres", "10.7", []string{"POSTGRES_USER=root", "POSTGRES_PASSWORD=secret", "POSTGRES_DB=test"})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err.Error())
		}

		// exponential backoff-retry, because the module in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			db, err = sql.Open("postgres", fmt.Sprintf("host=localhost port=%s user=root password=secret sslmode=disable database=test", resource.GetPort("5432/tcp")))
			//db, err = pgx.Connect(context.Background(),fmt.Sprintf("host=localhost port=%s user=root password=secret sslmode=disable database=test", resource.GetPort("5432/tcp")))
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			log.Fatalf("Could not connect to docker: %s", err.Error())
		}
		//setup gorm connection
		gormDB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), nil)
		if err != nil {
			log.Fatalf("failed to create postgresql database gorm connection '%s'", err)
		}
		code := m.Run()

		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err.Error())
		}

		os.Exit(code)
	case "sqlite3":
		db, err = sql.Open(*database, "test.db")
		if err != nil {
			log.Fatalf("failed to create sqlite database '%s'", err)
		}
		//setup gorm connection
		gormDB, err = gorm.Open(&sqlite.Dialector{
			Conn: db,
		}, nil)
		if err != nil {
			log.Fatalf("failed to create sqlite database gorm connection '%s'", err)
		}

		code := m.Run()

		os.Remove("test.db")
		os.Exit(code)
	case "mysql":
		log.Info("Started mysql database")
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}

		// pulls an image, creates a container based on it and runs it
		resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err = pool.Retry(func() error {
			db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql?sql_mode='ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION'&parseTime=%s", resource.GetPort("3306/tcp"), strconv.FormatBool(true)))
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}
		//setup gorm connection
		gormDB, err = gorm.Open(mysql.New(mysql.Config{
			Conn: db,
		}), nil)
		if err != nil {
			log.Fatalf("failed to create postgresql database gorm connection '%s'", err)
		}
		code := m.Run()
		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err.Error())
		}
		os.Exit(code)
	}

}

func TestEventRepositoryGorm_GetByAggregate(t *testing.T) {
	gormDB.Where("1 = 1").Unscoped().Delete(weos.GormEvent{})
	eventRepository, err := weos.NewBasicEventRepository(gormDB, log.New(), false, "123", "456")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}
	err = eventRepository.(*weos.EventRepositoryGorm).Migrate(context.Background())
	if err != nil {
		t.Fatalf("error encountered migration event repository '%s'", err)
	}
	entity := &weos.AggregateRoot{
		BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"},
	}

	mockEvent := weos.NewEntityEvent("CREATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title string `json:"title"`
	}{Title: "First Post"})

	mockEvent2 := weos.NewEntityEvent("UPDATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Lorem Ipsum"})

	mockEvent3 := weos.NewEntityEvent("UPDATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Finalizing Post"})

	entity.NewChange(mockEvent)
	entity.NewChange(mockEvent2)
	entity.NewChange(mockEvent3)

	err = eventRepository.Persist(context.TODO(), entity)
	if err != nil {
		t.Fatalf("error encountered persisting events '%s'", err)
	}

	events, err := eventRepository.GetByAggregate("1iNfR0jYD9UbYocH8D3WK6N4pG9")
	if err != nil {
		t.Fatalf("encountered error getting aggregate '%s' error: '%s'", "1iNfR0jYD9UbYocH8D3WK6N4pG9", err)
	}

	if len(events) != 3 {
		t.Errorf("expected %d events got %d", 3, len(events))
	}
}

func TestEventRepositoryGorm_GetByAggregateAndType(t *testing.T) {
	gormDB.Where("1 = 1").Unscoped().Delete(weos.GormEvent{})
	eventRepository, err := weos.NewBasicEventRepository(gormDB, log.New(), false, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}
	err = eventRepository.(*weos.EventRepositoryGorm).Migrate(context.Background())
	if err != nil {
		t.Fatalf("failed to run migrations")
	}

	entity := &weos.AggregateRoot{
		BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"},
	}

	mockEvent := weos.NewEntityEvent("CREATE_POST", entity, "1wqoyqIRsZTtnP3wjKh2Mq1Qp03", &struct {
		Title string `json:"title"`
	}{Title: "First Post"})

	mockEvent2 := weos.NewEntityEvent("UPDATE_POST", entity, "1wqoyqIRsZTtnP3wjKh2Mq1Qp03", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Lorem Ipsum"})

	mockEvent3 := weos.NewEntityEvent("UPDATE_POST", entity, "1wqoyqIRsZTtnP3wjKh2Mq1Qp03", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Finalizing Post"})

	entity.NewChange(mockEvent)
	entity.NewChange(mockEvent2)
	entity.NewChange(mockEvent3)

	err = eventRepository.Persist(context.TODO(), entity)
	if err != nil {
		t.Fatalf("error encountered persisting events '%s'", err)
	}

	events, err := eventRepository.GetByAggregateAndType("1iNfR0jYD9UbYocH8D3WK6N4pG9", "AggregateRoot")
	if err != nil {
		t.Fatalf("encountered error getting aggregate '%s' error: '%s'", "1iNfR0jYD9UbYocH8D3WK6N4pG9", err)
	}

	if len(events) != 3 {
		t.Errorf("expected %d events got %d", 3, len(events))
	}
}

func TestEventRepositoryGorm_GetByEntityAndAggregate(t *testing.T) {
	eventRepository, err := weos.NewBasicEventRepository(gormDB, log.New(), false, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}
	err = eventRepository.(*weos.EventRepositoryGorm).Migrate(context.Background())
	if err != nil {
		t.Fatalf("failed to run migrations")
	}
	entity := &weos.AggregateRoot{
		BasicEntity: weos.BasicEntity{ID: "1wqoyqIRsZTtnP3wjKh2Mq1Qp03"},
	}

	mockEvent := weos.NewEntityEvent("CREATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title string `json:"title"`
	}{Title: "First Post"})

	mockEvent2 := weos.NewEntityEvent("UPDATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Lorem Ipsum"})

	mockEvent3 := weos.NewEntityEvent("UPDATE_POST", entity, "1iNfR0jYD9UbYocH8D3WK6N4pG9", &struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}{Title: "Updated First Post", Description: "Finalizing Post"})

	entity.NewChange(mockEvent)
	entity.NewChange(mockEvent2)
	entity.NewChange(mockEvent3)

	err = eventRepository.Persist(context.TODO(), entity)
	if err != nil {
		t.Fatalf("error encountered persisting events '%s'", err)
	}

	events, err := eventRepository.GetByEntityAndAggregate("1wqoyqIRsZTtnP3wjKh2Mq1Qp03", "AggregateRoot", "1iNfR0jYD9UbYocH8D3WK6N4pG9")
	if err != nil {
		t.Fatalf("encountered error getting aggregate '%s' error: '%s'", "1iNfR0jYD9UbYocH8D3WK6N4pG9", err)
	}

	if len(events) != 3 {
		t.Errorf("expected %d events got %d", 3, len(events))
	}
}

func TestSaveAggregateEvents(t *testing.T) {
	type BaseAggregate struct {
		weos.AggregateRoot
		Title string `json:"title"`
	}

	baseAggregate := &BaseAggregate{}
	event, err := weos.NewBasicEvent("test.event", "123", "BaseAggregate", "")
	if err != nil {
		t.Fatalf("unexpected error setting up test event '%s'", err)
	}
	baseAggregate.NewChange(event)
	eventRepository, err := weos.NewBasicEventRepository(gormDB, log.New(), false, "123", "456")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}
	err = eventRepository.(*weos.EventRepositoryGorm).Migrate(context.Background())
	if err != nil {
		t.Fatalf("failed to run migrations")
	}
	err = eventRepository.Persist(context.TODO(), baseAggregate)
	if err != nil {
		t.Fatalf("encountered error persiting aggregate events")
	}

	if len(baseAggregate.GetNewChanges()) > 0 {
		t.Error("expected the list of new changes to be cleared")
	}
}

func TestEventRepositoryGorm_BatchPersist(t *testing.T) {

	eventRepository, err := weos.NewBasicEventRepository(gormDB, log.New(), false, "accountID", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}
	err = eventRepository.(*weos.EventRepositoryGorm).Migrate(context.Background())
	if err != nil {
		t.Fatalf("error setting up application'%s'", err)
	}

	generateEvents := make([]*weos.Event, 20000)
	entity := &weos.AggregateRoot{}

	for i := 0; i < 20000; i++ {

		currValue := strconv.Itoa(i)

		currEvent := "TEST_EVENT "
		currID := "batch id"
		currType := "SomeType"

		currEvent += currValue

		generateEvents[i] = &weos.Event{
			ID:      ksuid.New().String(),
			Type:    currEvent,
			Payload: nil,
			Meta: weos.EventMeta{
				EntityID:   currID,
				EntityType: currType,
				SequenceNo: 0,
			},
			Version: 1,
		}

		entity.NewChange(generateEvents[i])
	}

	//add an event handler
	eventHandlerCalled := 0
	eventRepository.AddSubscriber(func(ctx context.Context, event weos.Event) {
		eventHandlerCalled += 1
	})

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "12345"), entity)
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	if eventHandlerCalled != 20000 {
		t.Errorf("expected event handlers to be called %d time, called %d times", 20000, eventHandlerCalled)
	}

	//Struct for query results
	type QueryResults struct {
		entityID      string
		eventType     string
		accountID     string
		applicationID string
		sequenceNo    int
	}

	var rows *sql.Rows

	if *database == "mysql" {
		rows, err = db.Query("SELECT entity_id,type, root_id,application_id,sequence_no FROM gorm_events WHERE entity_id  = ? ORDER BY sequence_no ASC", "batch id")
		if err != nil {
			t.Fatalf("error retrieving events '%s'", err)
		}
	} else {
		rows, err = db.Query("SELECT entity_id,type, root_id,application_id FROM gorm_events WHERE entity_id  = $1", "batch id")
		if err != nil {
			t.Fatalf("error retrieving events '%s'", err)
		}
	}

	defer rows.Close()
	var rowInfo QueryResults
	var queryRows []QueryResults

	if *database == "mysql" {
		for rows.Next() {
			err = rows.Scan(&rowInfo.entityID, &rowInfo.eventType, &rowInfo.accountID, &rowInfo.applicationID, &rowInfo.sequenceNo)
			if err != nil {
				t.Fatalf("error retrieving event '%s'", err)
			}
			queryRows = append(queryRows, rowInfo)
		}
	} else {
		for rows.Next() {
			err = rows.Scan(&rowInfo.entityID, &rowInfo.eventType, &rowInfo.accountID, &rowInfo.applicationID)
			if err != nil {
				t.Fatalf("error retrieving event '%s'", err)
			}
			queryRows = append(queryRows, rowInfo)
		}
	}

	for i := 0; i < 20000; i++ {

		if queryRows[i].eventType != generateEvents[i].Type {
			t.Errorf("expected the type to be '%s', got '%s'", generateEvents[i].Type, queryRows[i].eventType)
		}

		if queryRows[i].accountID != "12345" {
			t.Errorf("expected the account id to be '%s', got '%s'", "12345", queryRows[i].accountID)
		}

		if queryRows[i].applicationID != "applicationID" {
			t.Errorf("expected the module id to be '%s', got '%s'", "applicationID", queryRows[i].applicationID)
		}
	}
}
