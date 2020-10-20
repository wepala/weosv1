package module

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos/errors"
	"github.com/wepala/weos/persistence"
	"net/http"
	"strconv"
)

type WeOSModule interface {
	GetModuleID() string
	GetTitle() string
	GetAccountID() string
	GetDBConnection() *sql.DB
	GetLogger() log.Ext1FieldLogger
	AddProjection(projection persistence.Projection) error
	GetProjections() []persistence.Projection
	Migrate(ctx context.Context) error
}

//Module is the core of the WeOS framework. It has a config, command handler and basic metadata as a default.
//This is a basic implementation and can be overwritten to include a db connection, httpCLient etc.
type WeOSMod struct {
	ModuleID          string `json:"moduleId"`
	Title             string `json:"title"`
	AccountID         string `json:"accountId"`
	commandDispatcher Dispatcher
	logger            log.Ext1FieldLogger
	db                *sql.DB
	httpClient        *http.Client
	projections       []persistence.Projection
}

func (w *WeOSMod) GetModuleID() string {
	return w.ModuleID
}

func (w *WeOSMod) GetTitle() string {
	return w.Title
}

func (w *WeOSMod) GetAccountID() string {
	return w.AccountID
}

func (w *WeOSMod) GetDBConnection() *sql.DB {
	return w.db
}

func (w *WeOSMod) GetLogger() log.Ext1FieldLogger {
	return w.logger
}

func (w *WeOSMod) AddProjection(projection persistence.Projection) error {
	w.projections = append(w.projections, projection)
	return nil
}

func (w *WeOSMod) GetProjections() []persistence.Projection {
	return w.projections
}

func (w *WeOSMod) Migrate(ctx context.Context) error {
	w.logger.WithField("module", w.Title).Infof("preparing to migrate %d projections", len(w.projections))
	for _, projection := range w.projections {
		err := projection.Migrate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

type WeOSModuleConfig struct {
	ModuleID  string         `json:"moduleId"`
	Title     string         `json:"title"`
	AccountID string         `json:"accountId"`
	Database  *WeOSDBConfig  `json:"database"`
	Log       *WeOSLogConfig `json:"log"`
}

type WeOSDBConfig struct {
	Host     string `json:"host"`
	User     string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Driver   string `json:"driver"`
	MaxOpen  int    `json:"max-open"`
	MaxIdle  int    `json:"max-idle"`
}

type WeOSLogConfig struct {
	Level        string `json:"level"`
	ReportCaller bool   `json:"report-caller"`
	Formatter    string `json:"formatter"`
}

//NewApplication creates a new basic module that allows for injecting of a few core components
//func NewApplication(applicationID string, applicationTitle string, accountID string, logger log.Ext1FieldLogger, db *sql.DB, httpClient *http.Client ) *WeOSMod {
//	return &WeOSMod{
//		ModuleID:    applicationID,
//		Title: applicationTitle,
//		AccountID:        accountID,
//		commandDispatcher:   &DefaultDispatcher{},
//		logger: logger,
//		db: db,
//		httpClient: httpClient,
//	}
//}

var NewApplicationFromConfig = func(config *WeOSModuleConfig, logger log.Ext1FieldLogger, db *sql.DB) (*WeOSMod, error) {

	var err error

	if logger == nil && config.Log != nil {
		if config.Log.Level != "" {
			switch config.Log.Level {
			case "debug":
				log.SetLevel(log.DebugLevel)
				break
			case "fatal":
				log.SetLevel(log.FatalLevel)
				break
			case "error":
				log.SetLevel(log.ErrorLevel)
				break
			case "warn":
				log.SetLevel(log.WarnLevel)
				break
			case "info":
				log.SetLevel(log.InfoLevel)
				break
			case "trace":
				log.SetLevel(log.TraceLevel)
				break
			}
		}

		if config.Log.Formatter == "json" {
			log.SetFormatter(&log.JSONFormatter{})
		}

		if config.Log.Formatter == "text" {
			log.SetFormatter(&log.TextFormatter{})
		}

		log.SetReportCaller(config.Log.ReportCaller)

		logger = log.New()
	}

	if db == nil && config.Database != nil {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.Database.Host, strconv.Itoa(config.Database.Port), config.Database.User, config.Database.Password, config.Database.Database)
		if config.Database.Driver == "" {
			config.Database.Driver = "postgres"
		}
		db, err = sql.Open(config.Database.Driver, connStr)
		if err != nil {
			return nil, errors.NewError("error setting up connection to database", err)
		}
		db.SetMaxOpenConns(config.Database.MaxOpen)
		db.SetMaxIdleConns(config.Database.MaxIdle)
	}

	return &WeOSMod{
		ModuleID:          config.ModuleID,
		Title:             config.Title,
		AccountID:         config.AccountID,
		commandDispatcher: &DefaultDispatcher{},
		logger:            logger,
		db:                db,
		httpClient:        nil,
	}, nil
}
