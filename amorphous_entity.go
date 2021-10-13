package weos

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const UI_SINGLE_LINE = "singleLine"
const UI_CHECKBOX = "checkbox"

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
	s.BasicProperty.UI = UI_SINGLE_LINE //Sets default

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
	b.BasicProperty.UI = UI_CHECKBOX //Sets default

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
	n.BasicProperty.UI = UI_SINGLE_LINE //Sets default

	return n
}

type AmorphousEntity struct {
	*BasicEntity
	Properties map[string]Property `json:"properties"`
}

func (e *AmorphousEntity) Get(label string) Property {
	return e.Properties[label]
}
func (e *AmorphousEntity) Set(property Property) {
	if e == nil {
		e = &AmorphousEntity{}
	}
	if e.Properties == nil {
		e.Properties = make(map[string]Property)
	}
	e.Properties[property.GetLabel()] = property
}

/*
//Umarshall AmorphousEntity into interface provided
func (e *AmorphousEntity) UnmarshalJSON(data []byte) error {

	for _, prop := range data {
		if string(prop) == "string" {
			stringProp := new(StringProperty).FromJSON(data)
			e.Properties[stringProp.GetLabel()] = stringProp
		}
		if string(prop) == "boolean" {
			booleanProp := new(BooleanProperty).FromJSON(data)
			e.Properties[booleanProp.GetLabel()] = booleanProp
		}
		if string(prop) == "numeric" {
			numericProp := new(NumericProperty).FromJSON(data)
			e.Properties[numericProp.GetLabel()] = numericProp
		}
	}

	return nil
}

func (s *StringProperty) FromJSON(prop []byte) *StringProperty {
	json.Unmarshal(prop, s)
	return s
}

func (b *BooleanProperty) FromJSON(prop []byte) *BooleanProperty {
	return nil
}

func (n *NumericProperty) FromJSON(prop []byte) *NumericProperty {
	return nil

}
*/

func (e *AmorphousEntity) UnmarshalJSON(data []byte) error {
	value, err := UnmarshalCustomValue(data, "id", "properties", map[string]reflect.Type{
		"string":  reflect.TypeOf(StringProperty{}),
		"boolean": reflect.TypeOf(BooleanProperty{}),
		"numeric": reflect.TypeOf(NumericProperty{}),
	})
	if err != nil {
		return err
	}

	e.Properties = value

	return nil

}

func UnmarshalCustomValue(data []byte, idJsonField, propertiesJsonField string, customTypes map[string]reflect.Type) (map[string]Property, error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	ID := m[idJsonField].(string)

	var value map[string]Property
	for i := 0; i <= len(m); i++ {
		if ty, found := customTypes[ID]; found {
			value = reflect.New(ty).Interface().(map[string]Property)
		}

		valueBytes, err := json.Marshal(m[propertiesJsonField])
		if err != nil {
			return nil, err
		}

		if propertiesJsonField == "string" {
			stringProp := new(StringProperty).FromJSON(data)
			value = stringProp
		}

	}

	return value, nil
}

func (s *StringProperty) FromJSON(prop []byte) *StringProperty {
	json.Unmarshal(prop, s)
	return s
}

func (b *BooleanProperty) FromJSON(prop []byte) *BooleanProperty {
	return nil
}

func (n *NumericProperty) FromJSON(prop []byte) *NumericProperty {
	return nil

}
