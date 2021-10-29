package integration_test

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"strconv"
	"testing"

	"github.com/go-redis/redis"
	"github.com/ory/dockertest/v3"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wepala/weos"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *sql.DB
var gormDB *gorm.DB
var database = flag.String("driver", "redis", "run database integration tests")
var err error
var rDatabase *redis.Client

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
	case "redis":
		log.Infof("Started redis")
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "redis",
			Name:         "redis-mock",
			ExposedPorts: []string{"6379:6379/tcp"},
			Env:          map[string]string{"REDIS_DB_URL": "redis:6379", "REDIS_DB_PASSWORD": "", "REDIS_DB": "0"},
			WaitingFor:   wait.ForLog("started"),
		}
		rContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start elastic search container '%s'", err)
		}

		defer rContainer.Terminate(ctx)

		//get the endpoint that the container was run on
		var endpoint string
		endpoint, err = rContainer.Host(ctx) //didn't use the endpoint call because it returns "localhost" which the client doesn't seem to like
		if err != nil {
			log.Fatalf("error setting up redis '%s'", err)
		}
		cport, err := rContainer.MappedPort(ctx, "6379")
		if err != nil {
			log.Fatalf("error setting up redis '%s'", err)
		}
		rEndpoint := endpoint + ":" + cport.Port()

		rDatabase = redis.NewClient(&redis.Options{
			Addr:     rEndpoint,
			Password: "",
			DB:       0,
		})
		pong, err := rDatabase.Ping().Result()
		if err != nil {
			panic(err)
		}

		if pong == "" {
			panic("no pong received")
		}

		code := m.Run()
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

func TestEventRepositoryRedis_GetByEntityAndAggregate(t *testing.T) {

	t.Run("get aggregate with 1 event ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(rDatabase, log.New(), "store", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}
		entity := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Item", RootID: "store", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "1iNfR0jYD9UbYocH8D3WK6N4pG9"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Post"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_ITEM",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Item", Description: "Shiny pearly necklace"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_ITEM",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 1,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Item", Description: "Finalizing Item"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_ITEM",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 3,
				RootID:     entity.RootID,
			},
			Version: 1,
		}

		entity.NewChange(mockEvent)
		entity.NewChange(mockEvent2)
		entity.NewChange(mockEvent3)

		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity.RootID), entity)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}

		events, err := eventRepository.GetByEntityAndAggregate(entity.ID, entity.Type, entity.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity.ID, err)
		}

		if len(events) != 3 {
			t.Errorf("expected %d events got %d", 3, len(events))
		}
		if events[0].ID != mockEvent.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent.ID, events[0].ID)
		}
		if events[0].Type != mockEvent.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
		}
		if events[1].ID != mockEvent2.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent2.ID, events[1].ID)
		}
		if events[1].Type != mockEvent2.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent2.Type, events[1].Type)
		}
		if events[2].ID != mockEvent3.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent3.ID, events[2].ID)
		}
		if events[2].Type != mockEvent3.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Type, events[2].Type)
		}

	})

	t.Run("get aggregate with 2 events with same type ", func(t *testing.T) {
		eventRepository, err := weos.NewRedisEventRepository(rDatabase, log.New(), "shop", "applicationID")
		if err != nil {
			t.Fatalf("error creating application '%s'", err)
		}

		entity := &struct {
			weos.AggregateRoot
			Type string `json:"type"`
		}{Type: "Snacks", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "another snack"}}}
		entity1 := &struct {
			weos.AggregateRoot
			Type   string `json:"type"`
			RootID string `json:"root_id"`
		}{Type: "Snacks", RootID: "shop", AggregateRoot: weos.AggregateRoot{
			BasicEntity: weos.BasicEntity{ID: "some snack"}}}

		payload, err := json.Marshal(&struct {
			Title string `json:"title"`
		}{Title: "First Snack"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_SNACK",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Updated First Snack", Description: "Lorem Ipsum"})

		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent2 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "UPDATE_SNACK",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 1,
			},
			Version: 1,
		}

		payload, err = json.Marshal(&struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}{Title: "Create Second Snack", Description: "Snack Snack Snack"})
		if err != nil {
			t.Fatalf("error encountered marshalling payload '%s'", err)
		}

		mockEvent3 := &weos.Event{
			ID:      ksuid.New().String(),
			Type:    "CREATE_SNACK",
			Payload: payload,
			Meta: weos.EventMeta{
				EntityID:   entity1.ID,
				EntityType: entity1.Type,
				SequenceNo: 0,
				RootID:     entity1.RootID,
			},
			Version: 1,
		}

		entity.NewChange(mockEvent)
		entity.NewChange(mockEvent2)
		entity1.NewChange(mockEvent3)

		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, "shop"), entity)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}
		err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity1.RootID), entity1)
		if err != nil {
			t.Fatalf("error encountered persisting event '%s'", err)
		}

		events, err := eventRepository.GetByEntityAndAggregate(entity.ID, entity.Type, "shop")
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity.ID, err)
		}

		if len(events) != 2 {
			t.Errorf("expected %d events got %d", 2, len(events))
		}
		if events[0].ID != mockEvent.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent.ID, events[0].ID)
		}
		if events[0].Type != mockEvent.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[0].Type)
		}
		if events[0].Meta.RootID != "shop" {
			t.Errorf("expected event type to be  '%s', got '%s'", "shop", events[0].Type)
		}
		if events[1].ID != mockEvent2.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent2.ID, events[1].ID)
		}
		if events[1].Type != mockEvent2.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent.Type, events[1].Type)
		}
		if events[1].Meta.RootID != "shop" {
			t.Errorf("expected event type to be  '%s', got '%s'", "shop", events[1].Type)
		}

		events, err = eventRepository.GetByEntityAndAggregate(entity1.ID, entity1.Type, entity1.RootID)
		if err != nil {
			t.Fatalf("encountered error getting aggregate '%s' error: '%s'", entity1.ID, err)
		}

		if len(events) != 1 {
			t.Errorf("expected %d events got %d", 1, len(events))
		}
		if events[0].ID != mockEvent3.ID {
			t.Errorf("expected event id to be  '%s', got '%s'", mockEvent3.ID, events[0].ID)
		}
		if events[0].Type != mockEvent3.Type {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Type, events[0].Type)
		}
		if events[0].Meta.RootID != mockEvent3.Meta.RootID {
			t.Errorf("expected event type to be  '%s', got '%s'", mockEvent3.Meta.RootID, events[0].Type)
		}
	})
}

func TestEventRepositoryRedis_BatchPersist(t *testing.T) {

	eventRepository, err := weos.NewRedisEventRepository(rDatabase, log.New(), "root123", "applicationID")
	if err != nil {
		t.Fatalf("error creating application '%s'", err)
	}

	generateEvents := make([]*weos.Event, 20000)
	entity := &struct {
		weos.AggregateRoot
		Type   string `json:"type"`
		RootID string `json:"root_id"`
	}{Type: "Post", RootID: "root123", AggregateRoot: weos.AggregateRoot{
		BasicEntity: weos.BasicEntity{ID: "batch id"}}}

	for i := 0; i < 20000; i++ {

		currValue := strconv.Itoa(i)
		currEvent := "TEST_EVENT "
		currEvent += currValue

		generateEvents[i] = &weos.Event{
			ID:      ksuid.New().String(),
			Type:    currEvent,
			Payload: nil,
			Meta: weos.EventMeta{
				EntityID:   entity.ID,
				EntityType: entity.Type,
				SequenceNo: 0,
				RootID:     entity.RootID,
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

	err = eventRepository.Persist(context.WithValue(context.TODO(), weos.ACCOUNT_ID, entity.RootID), entity)
	if err != nil {
		t.Fatalf("error encountered persisting event '%s'", err)
	}

	if eventHandlerCalled != 20000 {
		t.Errorf("expected event handlers to be called %d time, called %d times", 20000, eventHandlerCalled)
	}
	var events []weos.RedisEvent

	results := rDatabase.Get(entity.ID + ":" + entity.Type + ":" + entity.RootID)
	if results.Err() != nil {
		t.Fatalf("error encountered getting event '%s'", err)
	}

	err = json.Unmarshal([]byte(results.Val()), &events)

	if err != nil {
		t.Fatalf("error encountered unmarshalling events '%s'", err)
	}
	if len(events) != 20000 {
		t.Fatalf("expected events to have %d got %d", 20000, len(events))
	}

}
