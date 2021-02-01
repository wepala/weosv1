package module

import (
	"context"
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/wepala/weos/errors"
	"github.com/wepala/weos/persistence"
)

type Log interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

type WeOSModule interface {
	GetModuleID() string
	GetTitle() string
	GetAccountID() string
	GetDBConnection() *sql.DB
	GetLogger() Log
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
	logger            Log
	db                *sql.DB
	HttpClient        *http.Client
	projections       []persistence.Projection
	AccountURL        string `json:"accountURL"`
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

func (w *WeOSMod) GetLogger() Log {
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
	w.logger.Infof("preparing to migrate %d projections", len(w.projections))
	for _, projection := range w.projections {
		err := projection.Migrate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

type WeOSModuleConfig struct {
	ModuleID    string         `json:"moduleId"`
	Title       string         `json:"title"`
	AccountID   string         `json:"accountId"`
	AccountName string         `json:"accountName"`
	Database    *WeOSDBConfig  `json:"database"`
	Log         *WeOSLogConfig `json:"log"`
	BaseURL     string         `json:"baseURL"`
	LoginURL    string         `json:"loginURL"`
	GraphQLURL  string         `json:"graphQLURL"`
	SessionKey  string         `json:"sessionKey"`
	Secret      string         `json:"secret"`
	AccountURL  string         `json:"accountURL"`
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
//func NewApplication(applicationID string, applicationTitle string, accountID string, logger Log, db *sql.DB, HttpClient *http.Client ) *WeOSMod {
//	return &WeOSMod{
//		ModuleID:    applicationID,
//		Title: applicationTitle,
//		AccountID:        accountID,
//		commandDispatcher:   &DefaultDispatcher{},
//		logger: logger,
//		db: db,
//		HttpClient: HttpClient,
//	}
//}

var NewApplicationFromConfig = func(config *WeOSModuleConfig, logger Log, db *sql.DB) (*WeOSMod, error) {

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
		AccountURL:        config.AccountURL,
		HttpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}, nil
}
