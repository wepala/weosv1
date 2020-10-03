package application

import (
	"context"
	"time"
)

//Command is a common interface that all incoming requests should implement.
type Command struct {
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
	Execute  Execute
	Metadata CommandMetadata `json:"metadata"`
}

type CommandMetadata struct {
	Version       int64
	ExecutionDate *time.Time
	ApplicationID string
	AccountID     string
	UserID        string
}

type CommandHandler interface {
	Dispatch(context context.Context, command *Command) (*time.Time, error)
}

//DefaultCommandHandler is used to execute commands
type DefaultCommandHandler struct {
}

type Execute func(context context.Context) (*time.Time, error)

//Dispatches a command and can execute a command at a later date
func (ch *DefaultCommandHandler) Dispatch(context context.Context, command *Command) (*time.Time, error) {
	if command.Metadata.ExecutionDate == nil {
		return command.Execute(context)
	}
	//TODO save command to be executed later
	return nil, nil
}
