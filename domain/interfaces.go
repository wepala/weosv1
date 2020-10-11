package domain

//go:generate moq -out mocks_test.go -pkg domain_test . EventRepository
type WeOSEntity interface {
	Entity
	GetUser() User
	GetAccount() Account
	SetUser(User)
	SetAccount(Account)
}
type Entity interface {
	ValueObject
	GetID() string
}

type ValueObject interface {
	IsValid() bool
	AddError(err error)
	GetErrors() []error
}

type EventSourcedEntity interface {
	Entity
	NewChange(event *Event)
	GetNewChanges() []Entity
}

type Reducer func(initialState Entity, event Event, next Reducer) Entity
