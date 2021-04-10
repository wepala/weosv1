package weos

import (
	"sync"
)

type EventDisptacher struct {
	handlers        []EventHandler
	handlerPanicked bool
	dispatch        sync.Mutex
}

func (e *EventDisptacher) Dispatch(event Event) {
	//mutex helps keep state between routines
	e.dispatch.Lock()
	defer e.dispatch.Unlock()
	var wg sync.WaitGroup
	for i := 0; i < len(e.handlers); i++ {
		handler := e.handlers[i]
		wg.Add(1)
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

func (e *EventDisptacher) GetSubscribers() []EventHandler {
	return e.handlers
}

type EventHandler func(event Event)