package module

import (
	"context"
	"time"
)

//go:generate moq -out mocks_test.go -pkg application_test . Dispatcher
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

type Dispatcher interface {
	Dispatch(ctx context.Context, command *Command) error
}

//DefaultDispatcher is used to execute commands
type DefaultDispatcher struct {
}

//Dispatches a command and can execute a command at a later date
func (ch *DefaultDispatcher) Dispatch(context context.Context, command *Command) error {
	if command.Metadata.ExecutionDate == nil {
		return command.Execute(context)
	}
	//TODO save command to be executed later
	return nil
}

type Execute func(context context.Context) error
