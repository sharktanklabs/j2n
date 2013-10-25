package j2n_test

import (
	"encoding/json"
	"fmt"
	"github.com/ygt/j2n"
)

type DogData struct {
	Name     string                      `json:"name"`
	Overflow map[string]*json.RawMessage `json:"-"`
}

type Dog struct {
	DogData
}

func (d *Dog) UnmarshalJSON(data []byte) error {
	return j2n.UnmarshalJSON(data, &d.DogData)
}

func (d Dog) MarshalJSON() ([]byte, error) {
	return j2n.MarshalJSON(&d.DogData)
}

func ExampleUnmarshalJSON() {
	data := []byte(`{"age":2,"name":"Fido"}`)

	d := Dog{}
	json.Unmarshal(data, &d)

	fmt.Printf("%s is %s", d.Name, *d.Overflow["age"])

	// Output:
	// Fido is 2
}
