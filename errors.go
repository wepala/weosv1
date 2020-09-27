package weos

//goland:noinspection GoNameStartsWithPackageName
type WeOSError struct {
	message     string
	err         error
	application string
	accountID   string
}

func (e *WeOSError) Error() string {
	return e.message
}

func (e *WeOSError) Unwrap() error {
	return e.err
}

type DomainError struct {
	WeOSError
	EntityID   string
	EntityType string
}

type ErrorFactory struct {
	application string
	accountID   string
}

func (f *ErrorFactory) NewError(message string, err error) *WeOSError {
	return &WeOSError{
		message:     message,
		err:         err,
		application: f.application,
		accountID:   f.accountID,
	}
}

func (f *ErrorFactory) NewDomainError(message string, entityType string, entityID string, err error) *DomainError {
	return &DomainError{
		WeOSError:  *f.NewError(message, err),
		EntityID:   entityID,
		EntityType: entityType,
	}
}

var NewErrorFactory = func(applicationID string, accountID string) *ErrorFactory {
	return &ErrorFactory{
		application: applicationID,
		accountID:   accountID,
	}
}
