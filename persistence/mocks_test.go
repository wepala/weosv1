// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package persistence_test

import (
	"context"
	"github.com/wepala/weos/domain"
	"github.com/wepala/weos/persistence"
	"sync"
)

var (
	lockEventRepositoryMockAddSubscriber                  sync.RWMutex
	lockEventRepositoryMockFlush                          sync.RWMutex
	lockEventRepositoryMockGetByAggregate                 sync.RWMutex
	lockEventRepositoryMockGetByAggregateAndSequenceRange sync.RWMutex
	lockEventRepositoryMockPersist                        sync.RWMutex
	lockEventRepositoryMockRemove                         sync.RWMutex
)

// Ensure, that EventRepositoryMock does implement persistence.EventRepository.
// If this is not the case, regenerate this file with moq.
var _ persistence.EventRepository = &EventRepositoryMock{}

// EventRepositoryMock is a mock implementation of persistence.EventRepository.
//
//     func TestSomethingThatUsesEventRepository(t *testing.T) {
//
//         // make and configure a mocked persistence.EventRepository
//         mockedEventRepository := &EventRepositoryMock{
//             AddSubscriberFunc: func(handler persistence.EventHandler)  {
// 	               panic("mock out the AddSubscriber method")
//             },
//             FlushFunc: func() error {
// 	               panic("mock out the Flush method")
//             },
//             GetByAggregateFunc: func(ID string) ([]*domain.Event, error) {
// 	               panic("mock out the GetByAggregate method")
//             },
//             GetByAggregateAndSequenceRangeFunc: func(ID string, start int64, end int64) ([]*domain.Event, error) {
// 	               panic("mock out the GetByAggregateAndSequenceRange method")
//             },
//             PersistFunc: func(entities []domain.Entity) error {
// 	               panic("mock out the Persist method")
//             },
//             RemoveFunc: func(entities []domain.Entity) error {
// 	               panic("mock out the Remove method")
//             },
//         }
//
//         // use mockedEventRepository in code that requires persistence.EventRepository
//         // and then make assertions.
//
//     }
type EventRepositoryMock struct {
	// AddSubscriberFunc mocks the AddSubscriber method.
	AddSubscriberFunc func(handler persistence.EventHandler)

	// FlushFunc mocks the Flush method.
	FlushFunc func() error

	// GetByAggregateFunc mocks the GetByAggregate method.
	GetByAggregateFunc func(ID string) ([]*domain.Event, error)

	// GetByAggregateAndSequenceRangeFunc mocks the GetByAggregateAndSequenceRange method.
	GetByAggregateAndSequenceRangeFunc func(ID string, start int64, end int64) ([]*domain.Event, error)

	// PersistFunc mocks the Persist method.
	PersistFunc func(entities []domain.Entity) error

	// RemoveFunc mocks the Remove method.
	RemoveFunc func(entities []domain.Entity) error

	// calls tracks calls to the methods.
	calls struct {
		// AddSubscriber holds details about calls to the AddSubscriber method.
		AddSubscriber []struct {
			// Handler is the handler argument value.
			Handler persistence.EventHandler
		}
		// Flush holds details about calls to the Flush method.
		Flush []struct {
		}
		// GetByAggregate holds details about calls to the GetByAggregate method.
		GetByAggregate []struct {
			// ID is the ID argument value.
			ID string
		}
		// GetByAggregateAndSequenceRange holds details about calls to the GetByAggregateAndSequenceRange method.
		GetByAggregateAndSequenceRange []struct {
			// ID is the ID argument value.
			ID string
			// Start is the start argument value.
			Start int64
			// End is the end argument value.
			End int64
		}
		// Persist holds details about calls to the Persist method.
		Persist []struct {
			// Entities is the entities argument value.
			Entities []domain.Entity
		}
		// Remove holds details about calls to the Remove method.
		Remove []struct {
			// Entities is the entities argument value.
			Entities []domain.Entity
		}
	}
}

// AddSubscriber calls AddSubscriberFunc.
func (mock *EventRepositoryMock) AddSubscriber(handler persistence.EventHandler) {
	if mock.AddSubscriberFunc == nil {
		panic("EventRepositoryMock.AddSubscriberFunc: method is nil but EventRepository.AddSubscriber was just called")
	}
	callInfo := struct {
		Handler persistence.EventHandler
	}{
		Handler: handler,
	}
	lockEventRepositoryMockAddSubscriber.Lock()
	mock.calls.AddSubscriber = append(mock.calls.AddSubscriber, callInfo)
	lockEventRepositoryMockAddSubscriber.Unlock()
	mock.AddSubscriberFunc(handler)
}

// AddSubscriberCalls gets all the calls that were made to AddSubscriber.
// Check the length with:
//     len(mockedEventRepository.AddSubscriberCalls())
func (mock *EventRepositoryMock) AddSubscriberCalls() []struct {
	Handler persistence.EventHandler
} {
	var calls []struct {
		Handler persistence.EventHandler
	}
	lockEventRepositoryMockAddSubscriber.RLock()
	calls = mock.calls.AddSubscriber
	lockEventRepositoryMockAddSubscriber.RUnlock()
	return calls
}

// Flush calls FlushFunc.
func (mock *EventRepositoryMock) Flush() error {
	if mock.FlushFunc == nil {
		panic("EventRepositoryMock.FlushFunc: method is nil but EventRepository.Flush was just called")
	}
	callInfo := struct {
	}{}
	lockEventRepositoryMockFlush.Lock()
	mock.calls.Flush = append(mock.calls.Flush, callInfo)
	lockEventRepositoryMockFlush.Unlock()
	return mock.FlushFunc()
}

// FlushCalls gets all the calls that were made to Flush.
// Check the length with:
//     len(mockedEventRepository.FlushCalls())
func (mock *EventRepositoryMock) FlushCalls() []struct {
} {
	var calls []struct {
	}
	lockEventRepositoryMockFlush.RLock()
	calls = mock.calls.Flush
	lockEventRepositoryMockFlush.RUnlock()
	return calls
}

// GetByAggregate calls GetByAggregateFunc.
func (mock *EventRepositoryMock) GetByAggregate(ID string) ([]*domain.Event, error) {
	if mock.GetByAggregateFunc == nil {
		panic("EventRepositoryMock.GetByAggregateFunc: method is nil but EventRepository.GetByAggregate was just called")
	}
	callInfo := struct {
		ID string
	}{
		ID: ID,
	}
	lockEventRepositoryMockGetByAggregate.Lock()
	mock.calls.GetByAggregate = append(mock.calls.GetByAggregate, callInfo)
	lockEventRepositoryMockGetByAggregate.Unlock()
	return mock.GetByAggregateFunc(ID)
}

// GetByAggregateCalls gets all the calls that were made to GetByAggregate.
// Check the length with:
//     len(mockedEventRepository.GetByAggregateCalls())
func (mock *EventRepositoryMock) GetByAggregateCalls() []struct {
	ID string
} {
	var calls []struct {
		ID string
	}
	lockEventRepositoryMockGetByAggregate.RLock()
	calls = mock.calls.GetByAggregate
	lockEventRepositoryMockGetByAggregate.RUnlock()
	return calls
}

// GetByAggregateAndSequenceRange calls GetByAggregateAndSequenceRangeFunc.
func (mock *EventRepositoryMock) GetByAggregateAndSequenceRange(ID string, start int64, end int64) ([]*domain.Event, error) {
	if mock.GetByAggregateAndSequenceRangeFunc == nil {
		panic("EventRepositoryMock.GetByAggregateAndSequenceRangeFunc: method is nil but EventRepository.GetByAggregateAndSequenceRange was just called")
	}
	callInfo := struct {
		ID    string
		Start int64
		End   int64
	}{
		ID:    ID,
		Start: start,
		End:   end,
	}
	lockEventRepositoryMockGetByAggregateAndSequenceRange.Lock()
	mock.calls.GetByAggregateAndSequenceRange = append(mock.calls.GetByAggregateAndSequenceRange, callInfo)
	lockEventRepositoryMockGetByAggregateAndSequenceRange.Unlock()
	return mock.GetByAggregateAndSequenceRangeFunc(ID, start, end)
}

// GetByAggregateAndSequenceRangeCalls gets all the calls that were made to GetByAggregateAndSequenceRange.
// Check the length with:
//     len(mockedEventRepository.GetByAggregateAndSequenceRangeCalls())
func (mock *EventRepositoryMock) GetByAggregateAndSequenceRangeCalls() []struct {
	ID    string
	Start int64
	End   int64
} {
	var calls []struct {
		ID    string
		Start int64
		End   int64
	}
	lockEventRepositoryMockGetByAggregateAndSequenceRange.RLock()
	calls = mock.calls.GetByAggregateAndSequenceRange
	lockEventRepositoryMockGetByAggregateAndSequenceRange.RUnlock()
	return calls
}

// Persist calls PersistFunc.
func (mock *EventRepositoryMock) Persist(entities []domain.Entity) error {
	if mock.PersistFunc == nil {
		panic("EventRepositoryMock.PersistFunc: method is nil but EventRepository.Persist was just called")
	}
	callInfo := struct {
		Entities []domain.Entity
	}{
		Entities: entities,
	}
	lockEventRepositoryMockPersist.Lock()
	mock.calls.Persist = append(mock.calls.Persist, callInfo)
	lockEventRepositoryMockPersist.Unlock()
	return mock.PersistFunc(entities)
}

// PersistCalls gets all the calls that were made to Persist.
// Check the length with:
//     len(mockedEventRepository.PersistCalls())
func (mock *EventRepositoryMock) PersistCalls() []struct {
	Entities []domain.Entity
} {
	var calls []struct {
		Entities []domain.Entity
	}
	lockEventRepositoryMockPersist.RLock()
	calls = mock.calls.Persist
	lockEventRepositoryMockPersist.RUnlock()
	return calls
}

// Remove calls RemoveFunc.
func (mock *EventRepositoryMock) Remove(entities []domain.Entity) error {
	if mock.RemoveFunc == nil {
		panic("EventRepositoryMock.RemoveFunc: method is nil but EventRepository.Remove was just called")
	}
	callInfo := struct {
		Entities []domain.Entity
	}{
		Entities: entities,
	}
	lockEventRepositoryMockRemove.Lock()
	mock.calls.Remove = append(mock.calls.Remove, callInfo)
	lockEventRepositoryMockRemove.Unlock()
	return mock.RemoveFunc(entities)
}

// RemoveCalls gets all the calls that were made to Remove.
// Check the length with:
//     len(mockedEventRepository.RemoveCalls())
func (mock *EventRepositoryMock) RemoveCalls() []struct {
	Entities []domain.Entity
} {
	var calls []struct {
		Entities []domain.Entity
	}
	lockEventRepositoryMockRemove.RLock()
	calls = mock.calls.Remove
	lockEventRepositoryMockRemove.RUnlock()
	return calls
}

var (
	lockProjectionMockGetEventHandler sync.RWMutex
	lockProjectionMockMigrate         sync.RWMutex
)

// Ensure, that ProjectionMock does implement persistence.Projection.
// If this is not the case, regenerate this file with moq.
var _ persistence.Projection = &ProjectionMock{}

// ProjectionMock is a mock implementation of persistence.Projection.
//
//     func TestSomethingThatUsesProjection(t *testing.T) {
//
//         // make and configure a mocked persistence.Projection
//         mockedProjection := &ProjectionMock{
//             GetEventHandlerFunc: func() persistence.EventHandler {
// 	               panic("mock out the GetEventHandler method")
//             },
//             MigrateFunc: func(ctx context.Context) error {
// 	               panic("mock out the Migrate method")
//             },
//         }
//
//         // use mockedProjection in code that requires persistence.Projection
//         // and then make assertions.
//
//     }
type ProjectionMock struct {
	// GetEventHandlerFunc mocks the GetEventHandler method.
	GetEventHandlerFunc func() persistence.EventHandler

	// MigrateFunc mocks the Migrate method.
	MigrateFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// GetEventHandler holds details about calls to the GetEventHandler method.
		GetEventHandler []struct {
		}
		// Migrate holds details about calls to the Migrate method.
		Migrate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
}

// GetEventHandler calls GetEventHandlerFunc.
func (mock *ProjectionMock) GetEventHandler() persistence.EventHandler {
	if mock.GetEventHandlerFunc == nil {
		panic("ProjectionMock.GetEventHandlerFunc: method is nil but Projection.GetEventHandler was just called")
	}
	callInfo := struct {
	}{}
	lockProjectionMockGetEventHandler.Lock()
	mock.calls.GetEventHandler = append(mock.calls.GetEventHandler, callInfo)
	lockProjectionMockGetEventHandler.Unlock()
	return mock.GetEventHandlerFunc()
}

// GetEventHandlerCalls gets all the calls that were made to GetEventHandler.
// Check the length with:
//     len(mockedProjection.GetEventHandlerCalls())
func (mock *ProjectionMock) GetEventHandlerCalls() []struct {
} {
	var calls []struct {
	}
	lockProjectionMockGetEventHandler.RLock()
	calls = mock.calls.GetEventHandler
	lockProjectionMockGetEventHandler.RUnlock()
	return calls
}

// Migrate calls MigrateFunc.
func (mock *ProjectionMock) Migrate(ctx context.Context) error {
	if mock.MigrateFunc == nil {
		panic("ProjectionMock.MigrateFunc: method is nil but Projection.Migrate was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	lockProjectionMockMigrate.Lock()
	mock.calls.Migrate = append(mock.calls.Migrate, callInfo)
	lockProjectionMockMigrate.Unlock()
	return mock.MigrateFunc(ctx)
}

// MigrateCalls gets all the calls that were made to Migrate.
// Check the length with:
//     len(mockedProjection.MigrateCalls())
func (mock *ProjectionMock) MigrateCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	lockProjectionMockMigrate.RLock()
	calls = mock.calls.Migrate
	lockProjectionMockMigrate.RUnlock()
	return calls
}
