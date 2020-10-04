package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos/domain"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EventRepositoryGorm struct {
	DB              *gorm.DB
	gormDB          *gorm.DB
	eventDispatcher EventDisptacher
	logger          log.Ext1FieldLogger
	ctx             context.Context
}

type GormEvent struct {
	gorm.Model
	ID            string
	EntityID      string `gorm:"index"`
	Payload       datatypes.JSON
	Type          string `gorm:"index"`
	AccountID     string `gorm:"index"`
	ApplicationID string `gorm:"index"`
	User          string `gorm:"index"`
	SequenceNo    int64
}

//NewGormEvent converts a domain event to something that is a bit easier for Gorm to work with
func NewGormEvent(event *domain.Event) (GormEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return GormEvent{}, err
	}
	return GormEvent{
		ID:            event.ID,
		EntityID:      event.Meta.EntityID,
		Payload:       payload,
		Type:          event.Type,
		AccountID:     event.Meta.Account,
		ApplicationID: event.Meta.Application,
		User:          event.Meta.User,
		SequenceNo:    event.Meta.SequenceNo,
	}, nil
}

func (e *EventRepositoryGorm) Persist(entities []domain.Entity) error {
	if len(entities) == 0 {
		return nil
	} //didn't think it should barf if an empty list is passed
	var gormEvents []GormEvent
	savePointID := "s" + ksuid.New().String() //NOTE the save point can't start with a number
	e.logger.Infof("persisting %d events with save point %s", len(entities), savePointID)
	e.DB.SavePoint(savePointID)
	for _, event := range entities {
		if !event.IsValid() {
			for _, terr := range event.GetErrors() {
				e.logger.Errorf("error encountered persisting entity '%s', '%s'", event.(*domain.Event).Meta.EntityID, terr)
			}
			e.logger.Warnf("rolling back saving events to %s", savePointID)
			e.DB.RollbackTo(savePointID)
			return event.GetErrors()[0]
		}

		gormEvent, err := NewGormEvent(event.(*domain.Event))
		if err != nil {
			return err
		}
		gormEvents = append(gormEvents, gormEvent)
	}
	db := e.DB.Create(gormEvents)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (e *EventRepositoryGorm) GetByAggregate(ID string) ([]*domain.Event, error) {
	var events []GormEvent
	result := e.DB.Where("entity_id = ?", ID).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}

	var tevents []*domain.Event

	for _, event := range events {
		tevents = append(tevents, &domain.Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: event.Payload,
			Meta: domain.EventMeta{
				EntityID:    event.EntityID,
				Account:     event.AccountID,
				Application: event.ApplicationID,
				User:        event.User,
				SequenceNo:  event.SequenceNo,
			},
			Version: 0,
		})
	}
	return tevents, nil
}

func (e *EventRepositoryGorm) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*domain.Event, error) {
	var events []GormEvent
	result := e.DB.Where("entity_id = ? AND sequence_no >=? AND sequence_no <= ?", ID, start, end).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	var tevents []*domain.Event

	for _, event := range events {
		tevents = append(tevents, &domain.Event{
			ID:      event.ID,
			Type:    event.Type,
			Payload: event.Payload,
			Meta: domain.EventMeta{
				EntityID:    event.EntityID,
				Account:     event.AccountID,
				Application: event.ApplicationID,
				User:        event.User,
				SequenceNo:  event.SequenceNo,
			},
			Version: 0,
		})
	}
	return tevents, nil
}

func (e *EventRepositoryGorm) AddSubscriber(handler EventHandler) {
	e.eventDispatcher.AddSubscriber(handler)
}

func (e *EventRepositoryGorm) Migrate() error {
	event, err := NewGormEvent(&domain.Event{})
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

func (e *EventRepositoryGorm) Remove(entities []domain.Entity) error {

	savePointID := "s" + ksuid.New().String() //NOTE the save point can't start with a number
	e.logger.Infof("persisting %d events with save point %s", len(entities), savePointID)
	e.DB.SavePoint(savePointID)
	for _, event := range entities {
		gormEvent, err := NewGormEvent(event.(*domain.Event))
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

var NewEventRepositoryWithGORM = func(db *sql.DB, config *gorm.Config, useUnitOfWork bool, logger log.Ext1FieldLogger, ctx context.Context) (EventRepository, error) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), config)
	if err != nil {
		return nil, err
	}
	if useUnitOfWork {
		transaction := gormDB.Begin()
		return &EventRepositoryGorm{DB: transaction, gormDB: gormDB, logger: logger, ctx: ctx}, nil
	}
	return &EventRepositoryGorm{DB: gormDB}, nil
}
