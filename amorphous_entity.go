package weos

import (
	"fmt"
)

//Property interface that all fields should implement
type Property interface {
	IsValid() (bool, []error)
	GetType() string
	GetLabel() string
}

//BasicProperty is basic struct for a property
type BasicProperty struct {
	Type       string      `json:"type"`
	Label      string      `json:"label"`
	Value      interface{} `json:"value"`
	IsRequired bool        `json:"is_required"`
}

func (b *BasicProperty) GetType() string {
	return b.Type
}

func (b *BasicProperty) GetLabel() string {
	return b.Label
}

//StringProperty basic string property
type StringProperty struct {
	*BasicProperty
	Value string `json:"value"`
}

func (s *StringProperty) IsValid() (bool, []error) {
	if s.IsRequired && s.Value == "" {
		return false, []error{fmt.Errorf("'%s' is required", s.Label)}
	}

	return true, nil
}

type AmorphousEntity struct {
	*BasicEntity
	Properties map[string]Property `json:"properties"`
}

func (e *AmorphousEntity) Get(label string) Property {
	return e.Properties[label]
}

func (e *AmorphousEntity) Set(property Property) {
	e.Properties[property.GetLabel()] = property
}
