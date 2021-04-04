package weos

import (
	"context"
	"errors"
	"sync"
	"time"
)

//Command is a common interface that all incoming requests should implement.
type Command struct {
	Type     string          `json:"type"`
	Payload  interface{}     `json:"payload"`
	Metadata CommandMetadata `json:"metadata"`
}

type CommandMetadata struct {
	Version       int64
	ExecutionDate *time.Time
	UserID        string
}

type Dispatcher interface {
	Dispatch(ctx context.Context, command *Command) error
}

type DefaultCommandDispatcher struct {
	handlers        map[string][]CommandHandler
	handlerPanicked bool
	dispatch        sync.Mutex
}

func (e *DefaultCommandDispatcher) Dispatch(ctx context.Context, command *Command) error {
	//mutex helps keep state between routines
	e.dispatch.Lock()
	defer e.dispatch.Unlock()
	var wg sync.WaitGroup
	var err error
	if handlers, ok := e.handlers[command.Type]; ok {
		var allHandlers []CommandHandler
		//lets see if there are any global handlers and add those
		if globalHandlers, ok := e.handlers["*"]; ok {
			allHandlers = append(allHandlers, globalHandlers...)
		}
		//now lets add the specific command handlers
		allHandlers = append(allHandlers, handlers...)

		for i := 0; i < len(allHandlers); i++ {
			handler := allHandlers[i]
			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						e.handlerPanicked = true
						err = errors.New("handlers panicked")
					}
					wg.Done()
				}()
				err = handler(command)
			}()
		}

		wg.Wait()
	}

	return err
}

func (e *DefaultCommandDispatcher) AddSubscriber(command *Command, handler CommandHandler) {
	if e.handlers == nil {
		e.handlers = map[string][]CommandHandler{}
	}
	e.handlers[command.Type] = append(e.handlers[command.Type], handler)
}

func (e *DefaultCommandDispatcher) GetSubscribers() map[string][]CommandHandler {
	return e.handlers
}

type CommandHandler func(command *Command) error
