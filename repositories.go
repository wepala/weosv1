package weos

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/segmentio/ksuid"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EventRepositoryGorm struct {
	DB              *gorm.DB
	gormDB          *gorm.DB
	eventDispatcher EventDisptacher
	logger          Log
	ctx             context.Context
	unitOfWork      bool
	AccountID       string
	ApplicationID   string
	GroupID         string
	UserID          string
}

type GormEvent struct {
	gorm.Model
	ID            string
	EntityID      string `gorm:"index"`
	EntityType    string `gorm:"index"`
	Payload       datatypes.JSON
	Type          string `gorm:"index"`
	AccountID     string `gorm:"index"`
	ApplicationID string `gorm:"index"`
	User          string `gorm:"index"`
	SequenceNo    int64
}

//NewGormEvent converts a domain event to something that is a bit easier for Gorm to work with
func NewGormEvent(event *Event) (GormEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return GormEvent{}, err
	}

	return GormEvent{
		ID:            event.ID,
		EntityID:      event.Meta.EntityID,
		EntityType:    event.Meta.EntityType,
		Payload:       payload,
		Type:          event.Type,
		AccountID:     event.Meta.Account,
		ApplicationID: event.Meta.Module,
		User:          event.Meta.User,
		SequenceNo:    event.Meta.SequenceNo,
	}, nil
}

func (e *EventRepositoryGorm) Persist(entity AggregateInterface) error {
	//TODO use the information in the context to get account info, module info. //didn't think it should barf if an empty list is passed
	var gormEvents []GormEvent
	entities := entity.GetNewChanges()
	savePointID := "s" + ksuid.New().String() //NOTE the save point can't start with a number
	e.logger.Infof("persisting %d events with save point %s", len(entities), savePointID)
	if e.unitOfWork {
		e.DB.SavePoint(savePointID)
	}

	for _, entity := range entities {
		event := entity.(*Event)
		//let's fill in meta data if it's not already in the object
		if event.Meta.User == "" {
			event.Meta.User = e.UserID
		}
		if event.Meta.Account == "" {
			event.Meta.Account = e.AccountID
		}
		if event.Meta.Module == "" {
			event.Meta.Module = e.ApplicationID
		}
		if event.Meta.Group == "" {
			event.Meta.Group = e.GroupID
		}
		if !event.IsValid() {
			for _, terr := range event.GetErrors() {
				e.logger.Errorf("error encountered persisting entity '%s', '%s'", event.Meta.EntityID, terr)
			}
			if e.unitOfWork {
				e.logger.Debugf("rolling back saving events to %s", savePointID)
				e.DB.RollbackTo(savePointID)
			}

			return event.GetErrors()[0]
		}

		gormEvent, err := NewGormEvent(event)
		if err != nil {
			return err
		}
		gormEvents = append(gormEvents, gormEvent)
	}
	db := e.DB.Create(gormEvents)
	if db.Error != nil {
		return db.Error
	}

	//call persist on the aggregate root to clear the new changes array
	entity.Persist()

	for _, entity := range entities {
		e.eventDispatcher.Dispatch(*entity.(*Event))
	}
	return nil
}

func (e *EventRepositoryGorm) GetByAggregate(ID string) ([]*Event, error) {
	var events []GormEvent
	result := e.DB.Order("sequence_no asc").Where("entity_id = ?", ID).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}

	var tevents []*Event

	for _, event := range events {
		tevents = append(tevents, &Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta: EventMeta{
				EntityID:   event.EntityID,
				EntityType: event.EntityType,
				Account:    event.AccountID,
				Module:     event.ApplicationID,
				User:       event.User,
				SequenceNo: event.SequenceNo,
			},
			Version: 0,
		})
	}
	return tevents, nil
}

func (e *EventRepositoryGorm) GetByAggregateAndType(ID string, entityType string) ([]*Event, error) {
	var events []GormEvent
	result := e.DB.Order("sequence_no asc").Where("entity_id = ? AND entity_type = ?", ID, entityType).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}

	var tevents []*Event

	for _, event := range events {
		tevents = append(tevents, &Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta: EventMeta{
				EntityID:   event.EntityID,
				EntityType: event.EntityType,
				Account:    event.AccountID,
				Module:     event.ApplicationID,
				User:       event.User,
				SequenceNo: event.SequenceNo,
			},
			Version: 0,
		})
	}
	return tevents, nil
}

func (e *EventRepositoryGorm) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*Event, error) {
	var events []GormEvent
	result := e.DB.Order("sequence_no asc").Where("entity_id = ? AND sequence_no >=? AND sequence_no <= ?", ID, start, end).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	var tevents []*Event

	for _, event := range events {
		tevents = append(tevents, &Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: json.RawMessage(event.Payload),
			Meta: EventMeta{
				EntityID:   event.EntityID,
				EntityType: event.EntityType,
				Account:    event.AccountID,
				Module:     event.ApplicationID,
				User:       event.User,
				SequenceNo: event.SequenceNo,
			},
			Version: 0,
		})
	}
	return tevents, nil
}

func (e *EventRepositoryGorm) AddSubscriber(handler EventHandler) {
	e.eventDispatcher.AddSubscriber(handler)
}

func (e *EventRepositoryGorm) GetSubscribers() ([]EventHandler, error) {
	return e.eventDispatcher.GetSubscribers(), nil
}

func (e *EventRepositoryGorm) Migrate(ctx context.Context) error {
	event, err := NewGormEvent(&Event{})
	if err != nil {
		return err
	}
	err = e.DB.AutoMigrate(&event)
	if err != nil {
		return err
	}

	return nil
}

func (e *EventRepositoryGorm) Flush() error {
	err := e.DB.Commit().Error
	e.DB = e.gormDB.Begin()
	return err
}

func (e *EventRepositoryGorm) Remove(entities []Entity) error {

	savePointID := "s" + ksuid.New().String() //NOTE the save point can't start with a number
	e.logger.Infof("persisting %d events with save point %s", len(entities), savePointID)
	e.DB.SavePoint(savePointID)
	for _, event := range entities {
		gormEvent, err := NewGormEvent(event.(*Event))
		if err != nil {
			return err
		}
		db := e.DB.Delete(gormEvent)
		if db.Error != nil {
			e.DB.RollbackTo(savePointID)
			return db.Error
		}
	}

	return nil
}

//deprecated
var NewEventRepositoryWithGORM = func(db *sql.DB, config *gorm.Config, useUnitOfWork bool, logger Log, ctx context.Context, accountID string, applicationID string, userID string, groupID string) (EventRepository, error) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), config)
	if err != nil {
		return nil, err
	}
	if useUnitOfWork {
		transaction := gormDB.Begin()
		return &EventRepositoryGorm{DB: transaction, gormDB: gormDB, logger: logger, ctx: ctx, unitOfWork: useUnitOfWork, AccountID: accountID, GroupID: groupID, ApplicationID: applicationID, UserID: userID}, nil
	}
	return &EventRepositoryGorm{DB: gormDB, logger: logger, ctx: ctx, AccountID: accountID, GroupID: groupID, ApplicationID: applicationID, UserID: userID}, nil
}

func NewBasicEventRepository(gormDB *gorm.DB, logger Log, useUnitOfWork bool, accountID string, applicationID string) (EventRepository, error) {
	if useUnitOfWork {
		transaction := gormDB.Begin()
		return &EventRepositoryGorm{DB: transaction, gormDB: gormDB, logger: logger, unitOfWork: useUnitOfWork, AccountID: accountID, ApplicationID: applicationID}, nil
	}
	return &EventRepositoryGorm{DB: gormDB, logger: logger, ctx: context.Background(), AccountID: accountID, ApplicationID: applicationID}, nil
}
