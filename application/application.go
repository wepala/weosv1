package application

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/wepala/weos/errors"
	"net/http"
	"strconv"
)

type WeOSApplication interface {
	GetApplicationID() string
	GetApplicationTitle() string
	GetAccountID() string
	GetDBConnection() *sql.DB
	GetLogger() log.Ext1FieldLogger
}

//Application is the core of the WeOS framework. It has a config, command handler and basic metadata as a default.
//This is a basic implementation and can be overwritten to include a db connection, httpCLient etc.
type WeOSApp struct {
	ApplicationID     string `json:"applicationId"`
	ApplicationTitle  string `json:"applicationTitle"`
	AccountID         string `json:"accountId"`
	commandDispatcher Dispatcher
	logger            log.Ext1FieldLogger
	db                *sql.DB
	httpClient        *http.Client
}

func (w *WeOSApp) GetApplicationID() string {
	return w.ApplicationID
}

func (w *WeOSApp) GetApplicationTitle() string {
	return w.ApplicationTitle
}

func (w *WeOSApp) GetAccountID() string {
	return w.AccountID
}

func (w *WeOSApp) GetDBConnection() *sql.DB {
	return w.db
}

func (w *WeOSApp) GetLogger() log.Ext1FieldLogger {
	return w.logger
}

type WeOSApplicationConfig struct {
	ApplicationID    string         `json:"applicationID"`
	ApplicationTitle string         `json:"applicationTitle"`
	AccountID        string         `json:"accountId"`
	Database         *WeOSDBConfig  `json:"database"`
	Log              *WeOSLogConfig `json:"log"`
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

//NewApplication creates a new basic application that allows for injecting of a few core components
//func NewApplication(applicationID string, applicationTitle string, accountID string, logger log.Ext1FieldLogger, db *sql.DB, httpClient *http.Client ) *WeOSApp {
//	return &WeOSApp{
//		ApplicationID:    applicationID,
//		ApplicationTitle: applicationTitle,
//		AccountID:        accountID,
//		commandDispatcher:   &DefaultDispatcher{},
//		logger: logger,
//		db: db,
//		httpClient: httpClient,
//	}
//}

func NewApplicationFromConfig(config *WeOSApplicationConfig, logger log.Ext1FieldLogger, db *sql.DB) (*WeOSApp, error) {

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

	return &WeOSApp{
		ApplicationID:     config.ApplicationID,
		ApplicationTitle:  config.ApplicationTitle,
		AccountID:         config.AccountID,
		commandDispatcher: &DefaultDispatcher{},
		logger:            logger,
		db:                db,
		httpClient:        nil,
	}, nil
}
