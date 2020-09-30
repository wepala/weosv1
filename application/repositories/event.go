package repositories

import (
	"database/sql"
	"encoding/json"
	"github.com/wepala/weos/domain"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EventRepositoryGorm struct {
	DB              *gorm.DB
	eventDispatcher EventDisptacher
}

type GormEvent struct {
	gorm.Model
	EntityID      string `gorm:"index"`
	Payload       datatypes.JSON
	Type          string `gorm:"index"`
	AccountID     string `gorm:"index"`
	ApplicationID string `gorm:"index"`
	User          string `gorm:"index"`
	SequenceNo    int64
}

//NewGormEvent converts a domain event to something that is a bit easier for Gorm to work with
func NewGormEvent(event domain.Event) (GormEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return GormEvent{}, err
	}
	return GormEvent{
		EntityID:      event.Meta.ID,
		Payload:       payload,
		Type:          event.Type,
		AccountID:     event.Meta.Account,
		ApplicationID: event.Meta.Application,
		User:          event.Meta.User,
		SequenceNo:    event.Meta.SequenceNo,
	}, nil
}

func (e *EventRepositoryGorm) Persist(entities []domain.Event) error {
	var gormEvents []GormEvent
	for _, event := range entities {
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
	return nil
}

func (e *EventRepositoryGorm) GetByAggregate(ID string) ([]domain.Event, error) {
	var events []domain.Event
	result := e.DB.Where("entity_id = ?", ID).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (e *EventRepositoryGorm) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]domain.Event, error) {
	var events []domain.Event
	result := e.DB.Where("entity_id = ? AND sequence_no >=? AND sequence_no <= ?", ID, start, end).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (e *EventRepositoryGorm) AddSubscriber(handler EventHandler) {
	e.eventDispatcher.AddSubscriber(handler)
}

func (e *EventRepositoryGorm) Migrate() error {
	event, err := NewGormEvent(domain.Event{})
	if err != nil {
		return err
	}
	err = e.DB.AutoMigrate(&event)
	if err != nil {
		return err
	}

	return nil
}

func NewEventRepositoryWithGORM(db *sql.DB, config *gorm.Config) (*EventRepositoryGorm, error) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), config)
	if err != nil {
		return nil, err
	}
	return &EventRepositoryGorm{DB: gormDB}, nil
}
