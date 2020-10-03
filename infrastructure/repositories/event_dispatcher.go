package repositories

import (
	"github.com/wepala/weos/domain"
	"sync"
)

type EventDisptacher struct {
	handlers        []EventHandler
	handlerPanicked bool
	dispatch        sync.Mutex
}

func (e *EventDisptacher) Dispatch(event domain.Event) {
	//mutex helps keep state between routines
	e.dispatch.Lock()
	defer e.dispatch.Unlock()
	var wg sync.WaitGroup
	wg.Add(len(e.handlers))
	for _, handler := range e.handlers {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					e.handlerPanicked = true
				}
				wg.Done()
			}()
			handler(event)
		}()
	}

	wg.Wait()
}

func (e *EventDisptacher) AddSubscriber(handler EventHandler) {
	e.handlers = append(e.handlers, handler)
}

type EventHandler func(event domain.Event)
