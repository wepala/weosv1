package application

import "time"

type Command interface {
	Execute()
	GetAggregateID() string
	Delay() *time.Time
}

type CommandHandler struct {

}

func (ch *CommandHandler) Dispatch(command Command) {
	if command.Delay() == nil {
		command.Execute()
	}
	//TODO save command to be executed later
}