package weos

//go:generate moq -out mocks_test.go -pkg weos_test . EventRepository Projection Log Dispatcher

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ApplicationConfig struct {
	ModuleID    string     `json:"moduleId"`
	Title       string     `json:"title"`
	AccountID   string     `json:"accountId"`
	AccountName string     `json:"accountName"`
	Database    *DBConfig  `json:"database"`
	Log         *LogConfig `json:"log"`
	BaseURL     string     `json:"baseURL"`
	LoginURL    string     `json:"loginURL"`
	GraphQLURL  string     `json:"graphQLURL"`
	SessionKey  string     `json:"sessionKey"`
	Secret      string     `json:"secret"`
	AccountURL  string     `json:"accountURL"`
}

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Driver   string `json:"driver"`
	MaxOpen  int    `json:"max-open"`
	MaxIdle  int    `json:"max-idle"`
}

type LogConfig struct {
	Level        string `json:"level"`
	ReportCaller bool   `json:"report-caller"`
	Formatter    string `json:"formatter"`
}

type Application interface {
	ID() string
	Title() string
	DBConnection() *sql.DB
	Logger() Log
	AddProjection(projection Projection) error
	Projections() []Projection
	Migrate(ctx context.Context) error
	Config() *ApplicationConfig
	EventRepository() EventRepository
	HTTPClient() *http.Client
}

//Module is the core of the WeOS framework. It has a config, command handler and basic metadata as a default.
//This is a basic implementation and can be overwritten to include a db connection, httpCLient etc.
type BaseApplication struct {
	id              string
	title           string
	logger          Log
	db              *sql.DB
	config          *ApplicationConfig
	projections     []Projection
	eventRepository EventRepository
	httpClient      *http.Client
}

func (w *BaseApplication) Logger() Log {
	return w.logger
}

func (w *BaseApplication) Config() *ApplicationConfig {
	return w.config
}

func (w *BaseApplication) ID() string {
	return w.id
}

func (w *BaseApplication) Title() string {
	return w.title
}

func (w *BaseApplication) DBConnection() *sql.DB {
	return w.db
}

func (w *BaseApplication) AddProjection(projection Projection) error {
	w.projections = append(w.projections, projection)
	return nil
}

func (w *BaseApplication) Projections() []Projection {
	return w.projections
}

func (w *BaseApplication) Migrate(ctx context.Context) error {
	w.logger.Infof("preparing to migrate %d projections", len(w.projections))
	for _, projection := range w.projections {
		err := projection.Migrate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *BaseApplication) EventRepository() EventRepository {
	return w.eventRepository
}

func (w *BaseApplication) HTTPClient() *http.Client {
	return w.httpClient
}

var NewApplicationFromConfig = func(config *ApplicationConfig, logger Log, db *sql.DB, client *http.Client, eventRepository EventRepository) (*BaseApplication, error) {

	var err error

	if logger == nil && config.Log != nil {
		//	if config.Log.Level != "" {
		//		switch config.Log.Level {
		//		case "debug":
		//			log.SetLevel(log.DebugLevel)
		//			break
		//		case "fatal":
		//			log.SetLevel(log.FatalLevel)
		//			break
		//		case "error":
		//			log.SetLevel(log.ErrorLevel)
		//			break
		//		case "warn":
		//			log.SetLevel(log.WarnLevel)
		//			break
		//		case "info":
		//			log.SetLevel(log.InfoLevel)
		//			break
		//		case "trace":
		//			log.SetLevel(log.TraceLevel)
		//			break
		//		}
		//	}
		//
		//	if config.Log.Formatter == "json" {
		//		log.SetFormatter(&log.JSONFormatter{})
		//	}
		//
		//	if config.Log.Formatter == "text" {
		//		log.SetFormatter(&log.TextFormatter{})
		//	}
		//
		//	log.SetReportCaller(config.Log.ReportCaller)
		//
		logger = log.New()
	}

	if db == nil && config.Database != nil {
		var connStr string

		if config.Database.Driver == "" {
			config.Database.Driver = "postgres"
		}

		switch config.Database.Driver {
		case "sqlite3":
			//check if file exists and if not create it. We only do this if a memory only db is NOT asked for
			//(Note that if it's a combination we go ahead and create the file) https://www.sqlite.org/inmemorydb.html
			if config.Database.Database != ":memory:" {
				if _, err = os.Stat(config.Database.Database); os.IsNotExist(err) {
					_, err = os.Create(strings.Replace(config.Database.Database, ":memory:", "", -1))
					if err != nil {
						return nil, NewError(fmt.Sprintf("error creating sqlite database '%s'", config.Database.Database), err)
					}
				}
			}

			connStr = fmt.Sprintf("%s",
				config.Database.Database)

			//update connection string to include authentication IF a username is set
			if config.Database.User != "" {
				authEnticationString := fmt.Sprintf("?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha512",
					config.Database.User, config.Database.Password)
				connStr = connStr + authEnticationString
			}
		default:
			connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				config.Database.Host, strconv.Itoa(config.Database.Port), config.Database.User, config.Database.Password, config.Database.Database)
		}

		db, err = sql.Open(config.Database.Driver, connStr)
		if err != nil {
			return nil, NewError("error setting up connection to database", err)
		}

		db.SetMaxOpenConns(config.Database.MaxOpen)
		db.SetMaxIdleConns(config.Database.MaxIdle)
	}

	if client == nil {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &BaseApplication{
		id:              config.ModuleID,
		title:           config.Title,
		logger:          logger,
		db:              db,
		config:          config,
		httpClient:      client,
		eventRepository: eventRepository,
	}, nil
}
