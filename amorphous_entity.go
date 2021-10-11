package weos

import "fmt"

//Property interface that all fields should implement
type Property interface {
	IsValid() bool
	GetType() string
	GetLabel() string
	GetErrors() []error
}

//BasicProperty is basic struct for a property
type BasicProperty struct {
	Type       string      `json:"type"`
	UI         string      `json:"ui"`
	Label      string      `json:"label"`
	Value      interface{} `json:"value"`
	IsRequired bool        `json:"is_required"`
	errors     []error
}

func (b *BasicProperty) GetType() string {
	return b.Type
}
func (b *BasicProperty) GetLabel() string {
	return b.Label
}
func (b *BasicProperty) GetErrors() []error {
	return b.errors
}

//StringProperty basic string property
type StringProperty struct {
	*BasicProperty
	Value string `json:"value"`
}

//IsValid add rules for validating value
func (s *StringProperty) IsValid() bool {
	if s.IsRequired && s.Value == "" {
		s.errors = append(s.errors, fmt.Errorf("'%s' is required", s.Label))
		return false
	}
	return true
}

//FromLabelAndValue create property using label
func (s *StringProperty) FromLabelAndValue(label string, value string, isRequired bool) *StringProperty {
	if s.BasicProperty == nil {
		s.BasicProperty = &BasicProperty{}
	}
	s.BasicProperty.Type = "string"
	s.BasicProperty.Label = label
	s.Value = value
	s.BasicProperty.IsRequired = isRequired
	return s
}

//BooleanProperty basic string property
type BooleanProperty struct {
	*BasicProperty
	Value bool `json:"value"`
}

//IsValid add rules for validating value
func (b *BooleanProperty) IsValid() bool {
	return true
}

//FromLabelAndValue create property using label
func (b *BooleanProperty) FromLabelAndValue(label string, value bool, isRequired bool) *BooleanProperty {
	if b.BasicProperty == nil {
		b.BasicProperty = &BasicProperty{}
	}
	b.BasicProperty.Type = "boolean"
	b.BasicProperty.Label = label
	b.Value = value
	b.BasicProperty.IsRequired = isRequired
	return b
}

//NumericProperty basic string property
type NumericProperty struct {
	*BasicProperty
	Value float32 `json:"value"`
}

//IsValid add rules for validating value
func (n *NumericProperty) IsValid() bool {
	if n.IsRequired && n.Value == 0 {
		n.errors = append(n.errors, fmt.Errorf("'%s' is required", n.Label))
		return false
	}
	return true
}

//FromLabelAndValue create property using label
func (n *NumericProperty) FromLabelAndValue(label string, value float32, isRequired bool) *NumericProperty {
	if n.BasicProperty == nil {
		n.BasicProperty = &BasicProperty{}
	}
	n.BasicProperty.Type = "numeric"
	n.BasicProperty.Label = label
	n.Value = value
	n.BasicProperty.IsRequired = isRequired
	return n
}

type AmorphousEntity struct {
	*BasicEntity
	properties map[string]Property `json:"properties"`
}

func (e *AmorphousEntity) Get(label string) Property {
	return e.properties[label]
}
func (e *AmorphousEntity) Set(property Property) {
	if e == nil {
		e = &AmorphousEntity{}
	}
	if e.properties == nil {
		e.properties = make(map[string]Property)
	}
	e.properties[property.GetLabel()] = property
}
