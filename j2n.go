// Package j2n allows arbitrary JSON to be marshaled into structs. Any JSON
// fields that are not marshaled directly into the fields of the struct are put
// into a field called 'Overflow', of type
//
// 	map[string]*json.RawMessage
//
// This means that fields that are not explicitly named in the struct will
// survive an Unmarshal/Marshal round trip.
//
// To avoid recursive calls to MarshalJSON/UnmarshalJSON, use the following
// pattern:
//
// 	type CatData struct {
// 		Name     string                      `json:"name"`
// 		Overflow map[string]*json.RawMessage `json:"-"`
// 	}
//
// 	type Cat struct {
// 		CatData
// 	}
//
// 	func (c *Cat) UnmarshalJSON(data []byte) error {
// 		return j2n.UnmarshalJSON(data, &c.CatData)
// 	}
//
// 	func (c Cat) MarshalJSON() ([]byte, error) {
// 		return j2n.MarshalJSON(c.CatData)
// 	}
//
package j2n

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Parses the JSON-encoded data into the struct pointed to by v.
//
// This behaves exactly like json.Unmarshal, but any extra JSON fields that
// are not explicitly named in the struct are unmarshaled in the 'Overflow'
// field.
//
// The struct v must contain a field 'Overflow' of type
//
//	map[string]*json.RawMessage
//
func UnmarshalJSON(data []byte, v interface{}) error {
	overflow, err := resetOverflowMap(v)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &overflow); err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	namedFieldsJSON, err := json.Marshal(v)
	if err != nil {
		return err
	}

	namedFieldsMap := make(map[string]*json.RawMessage)
	if err := json.Unmarshal(namedFieldsJSON, &namedFieldsMap); err != nil {
		return err
	}

	for k, _ := range namedFieldsMap {
		delete(overflow, k)
	}

	return nil
}

// Returns the JSON encoding of v, which must be a struct.
//
// This behaves exactly like json.Marshal, but ensures that any extra fields
// mentioned in v.Overflow are output alongside the explicitly named struct
// fields.
//
// It expects v to contain a field named 'Overflow' of type
//
// 	map[string]*json.RawMessage
//
func MarshalJSON(v interface{}) ([]byte, error) {
	result := make(map[string]*json.RawMessage)

	// Do a round trip of the named fields into a map[string]*json.RawMessage
	namedFieldsJSON, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(namedFieldsJSON, &result)
	if err != nil {
		return nil, err
	}

	overflow, err := getOverflowMap(v)
	if err != nil {
		return nil, err
	}

	for k, v := range overflow {
		if _, ok := result[k]; ok {
			errorText := fmt.Sprintf("Named field present in overflow: '%s'", k)
			return nil, errors.New(errorText)
		}
		result[k] = v
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}

func resetOverflowMap(v interface{}) (map[string]*json.RawMessage, error) {
	if value, err := getOverflowFieldValue(v); err != nil {
		return nil, err
	} else {
		overflow := make(map[string]*json.RawMessage)
		value.Set(reflect.ValueOf(overflow))
		return overflow, nil
	}
}

func getOverflowMap(v interface{}) (map[string]*json.RawMessage, error) {
	if value, err := getOverflowFieldValue(v); err != nil {
		return nil, err
	} else {
		return value.Interface().(map[string]*json.RawMessage), nil
	}
}

func getOverflowFieldValue(v interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(v)

	// Unwrap the pointer if necessary
	if value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Check that we're dealing with a struct
	if value.Type().Kind() != reflect.Struct {
		errText := fmt.Sprintf("Expected struct, got %s", value.Type().Kind())
		return reflect.Value{}, errors.New(errText)
	}

	// Ensure the struct has a field called 'Overflow'
	overflowField := value.FieldByName("Overflow")
	if !overflowField.IsValid() {
		return reflect.Value{}, errors.New("Overflow field is missing")
	}

	// And that the field has type map[string]*json.RawMessage
	if overflowField.Type() != reflect.TypeOf(make(map[string]*json.RawMessage)) {
		return reflect.Value{}, errors.New("Overflow must be of type map[string]*json.RawMessage")
	}

	// And that it has a tag ensuring that it is omitted from the JSON output
	overflowFieldType, _ := value.Type().FieldByName("Overflow")
	if overflowFieldType.Tag != `json:"-"` {
		return reflect.Value{}, errors.New("Overflow must be of type map[string]*json.RawMessage")
	}

	return overflowField, nil
}
