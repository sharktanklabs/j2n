package j2n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

type NonStruct int

func (n *NonStruct) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(data, n)
}

func TestReturnsErrorUnmarshalingToNonStructType(t *testing.T) {
	var n NonStruct

	err := json.Unmarshal([]byte(`{}`), &n)
	if err == nil {
		t.Fatal("Expected error unmarshaling into non-struct type")
	}
}

type PersonDataWithoutOverflow struct {
	Name string `json:"name"`
}

type PersonWithoutOverflow struct {
	PersonDataWithoutOverflow
}

func (p *PersonWithoutOverflow) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(data, &p.PersonDataWithoutOverflow)
}

func TestReturnsErrorWhenOverflowFieldIsMissing(t *testing.T) {
	p := PersonWithoutOverflow{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), &p)
	if err == nil {
		t.Fatal("Expected error unmarshaling into struct with no Overflow field")
	}
}

type PersonDataWithIncorrectOverflow struct {
	Name     string `json:"name"`
	Overflow string
}

type PersonWithIncorrectOverflow struct {
	PersonDataWithIncorrectOverflow
}

func (p *PersonWithIncorrectOverflow) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(data, &p.PersonDataWithIncorrectOverflow)
}

func TestReturnsErrorWithIncorrectOverflowFieldType(t *testing.T) {
	p := PersonWithIncorrectOverflow{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), &p)
	if err == nil {
		t.Fatal("Expected error unmarshaling into struct with Overflow field of incorrect type")
	}
}

type PersonDataWithoutOverflowTag struct {
	Name     string `json:"name"`
	Overflow map[string]*json.RawMessage
}

type PersonWithoutOverflowTag struct {
	PersonDataWithoutOverflowTag
}

func (p *PersonWithoutOverflowTag) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(data, &p.PersonDataWithoutOverflowTag)
}

func TestReturnsErrorWhenOverflowTagMissing(t *testing.T) {
	p := PersonWithoutOverflowTag{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), &p)
	if err == nil {
		t.Fatal("Expected error with Overflow field missing `json:\"-\"` tag")
	}
}

type PersonData struct {
	Name     string                      `json:"name"`
	Overflow map[string]*json.RawMessage `json:"-"`
}

type Person struct {
	PersonData
}

func (p *Person) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(data, &p.PersonData)
}

func (p Person) MarshalJSON() ([]byte, error) {
	return MarshalJSON(&p.PersonData)
}

func TestReturnsErrorUnmarshalingToNonPointerType(t *testing.T) {
	p := Person{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), p)
	if err == nil {
		t.Fatal("Expected error unmarshaling into non-pointer type")
	}
}

func TestParsesNamedFields(t *testing.T) {
	p := Person{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), &p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	expectedName := "Bert"
	if p.Name != expectedName {
		t.Fatalf("Expected '%s', got '%s'", expectedName, p.Name)
	}
}

func TestParsesOverflowFields(t *testing.T) {
	p := Person{}

	err := json.Unmarshal([]byte(`{"age":29}`), &p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	if p.Overflow == nil {
		t.Fatal("Overflow was nil")
	}

	actualAgeJSON, ok := p.Overflow["age"]
	if !ok {
		t.Fatal("'age' field in Overflow was missing")
	}

	expectedAgeJSON := json.RawMessage(`29`)
	if !bytes.Equal(*actualAgeJSON, expectedAgeJSON) {
		t.Fatalf("Expected '%s', got '%s'", expectedAgeJSON, p.Overflow["age"])
	}
}

func TestCopiesOverflowFieldsVerbatim(t *testing.T) {
	p := Person{}

	fieldJSON := []byte(`{"bar": 99999999999999999999}`)
	outerJSON := []byte(fmt.Sprintf(`{"foo":%s}`, fieldJSON))

	err := json.Unmarshal(outerJSON, &p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	if !bytes.Equal(fieldJSON, *p.Overflow["foo"]) {
		t.Fatalf("Expected '%s', got '%s'", fieldJSON, *p.Overflow["foo"])
	}
}

func TestNamedFieldsAreNotInOverflow(t *testing.T) {
	p := Person{}

	err := json.Unmarshal([]byte(`{"name":"Bert"}`), &p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	if p.Overflow["name"] != nil {
		t.Fatalf("Expected 'name' to be absent from Overflow, got '%s'", p.Overflow["name"])
	}
}

func TestMarshaledOutputContainsNamedFields(t *testing.T) {
	p := Person{}
	p.Name = "Bert"

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	expectedData := []byte(`{"name":"Bert"}`)
	if !bytes.Equal(data, expectedData) {
		t.Fatalf("Expected '%s', got '%s'", expectedData, data)
	}
}

func TestMarshaledOutputContainsOverflowFields(t *testing.T) {
	p := Person{}
	p.Overflow = make(map[string]*json.RawMessage)

	ageJSON := json.RawMessage("29")
	p.Overflow["age"] = &ageJSON

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}

	expectedData := []byte(`{"age":29,"name":""}`)
	if !bytes.Equal(data, expectedData) {
		t.Fatalf("Expected '%s', got '%s'", expectedData, data)
	}
}

func TestErrorOnAliasedFields(t *testing.T) {
	p := Person{}
	p.Overflow = make(map[string]*json.RawMessage)

	nameJSON := json.RawMessage(`"Bert"`)
	p.Overflow["name"] = &nameJSON

	_, err := json.Marshal(p)
	if err == nil {
		t.Fatal("Expected error on aliased fields, got none")
	}
}
