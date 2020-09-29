package application

import (
	"github.com/spf13/viper"
	"time"
)

//Application is the core of the WeOS framework. It has a config, command handler and basic metadata as a default.
//This is a basic implementation and can be overwritten to include a db connection, httpCLient etc.
type Application struct {
	ApplicationId    string `json:"applicationId"`
	ApplicationTitle string `json:"applicationTitle"`
	AccountId        string `json:"accountId"`
	Config           *viper.Viper
	commandHandler   CommandHandler
}

func (app *Application) Run(command Command) (*time.Time, error) {
	return app.commandHandler.Dispatch(command)
}

//NewApplication creates a new basic application.
func NewApplication(config *viper.Viper) *Application {
	return &Application{
		ApplicationId:    config.GetString("APPLICATION_ID"),
		ApplicationTitle: config.GetString("APPLICATION_TITLE"),
		AccountId:        config.GetString("ACCOUNT_ID"),
		commandHandler:   CommandHandler{},
	}
}
