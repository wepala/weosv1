package domain

import (
	"encoding/json"
	"github.com/wepala/weos/errors"
)

//AggregateRoot Is a base struct for WeOS applications to use. This is event sourcing ready by default
type AggregateRoot struct {
	BasicEntity
	SequenceNo int64
	newEvents  []Entity
	User       User
	Account    Account
}

func (w *AggregateRoot) GetUser() User {
	return w.User
}

func (w *AggregateRoot) GetAccount() Account {
	return w.Account
}

func (w *AggregateRoot) SetUser(user User) {
	w.User = user
}

func (w *AggregateRoot) SetAccount(account Account) {
	w.Account = account
}

func (w *AggregateRoot) NewChange(event *Event) {
	w.newEvents = append(w.newEvents, event)
}

func (w *AggregateRoot) GetNewChanges() []Entity {
	return w.newEvents
}

var DefaultReducer = func(initialState Entity, event *Event, next Reducer) Entity {
	//convert event to json string
	eventString, err := json.Marshal(event.Payload)
	if err != nil {
		initialState.AddError(errors.NewDomainError("error marshalling event", "", initialState.GetID(), err))
	} else {
		err := json.Unmarshal(eventString, &initialState)
		if err != nil {
			initialState.AddError(errors.NewDomainError("error unmarshalling event into entity", "", initialState.GetID(), err))
		}
	}
	//if it's an aggregate root then let's set the user and account based on the event meta details
	if aggregateRoot, ok := initialState.(WeOSEntity); ok {
		aggregateRoot.SetUser(User{
			BasicEntity{
				ID: event.Meta.User,
			},
		})
		aggregateRoot.SetAccount(Account{
			BasicEntity{
				ID: event.Meta.Account,
			},
		})
	}

	return initialState
}

var NewAggregateFromEvents = func(initialState Entity, events []*Event) Entity {
	for _, event := range events {
		initialState = DefaultReducer(initialState, event, nil)
	}

	return initialState
}
