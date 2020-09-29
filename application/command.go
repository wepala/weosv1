package application

import "time"

//Command is a common interface that all incoming requests should implement.
type Command interface {
	Execute() (*time.Time, error)
	GetPayload() interface{}
	GetType() string
	GetMetadata() CommandMetadata
}

type CommandMetadata struct {
	Version       int64
	ExecutionDate *time.Time
	ApplicationID string
	AccountID     string
	UserID        string
}

//CommandHandler is used to execute commands
type CommandHandler struct {
}

//Dispatches a command and can execute a command at a later date
func (ch *CommandHandler) Dispatch(command Command) (*time.Time, error) {
	if command.GetMetadata().ExecutionDate == nil {
		return command.Execute()
	}
	//TODO save command to be executed later
	return nil, nil
}
