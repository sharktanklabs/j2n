package j2n_test

import (
	"encoding/json"
	"fmt"
	"github.com/ygt/j2n"
)

type CatData struct {
	Name     string                      `json:"name"`
	Overflow map[string]*json.RawMessage `json:"-"`
}

type Cat struct {
	CatData
}

func (c *Cat) UnmarshalJSON(data []byte) error {
	return j2n.UnmarshalJSON(data, &c.CatData)
}

func (c Cat) MarshalJSON() ([]byte, error) {
	return j2n.MarshalJSON(&c.CatData)
}

func ExampleMarshalJSON() {
	c := Cat{}
	c.Name = "Tiddles"
	c.Overflow = make(map[string]*json.RawMessage)

	age := json.RawMessage("2")
	c.Overflow["age"] = &age

	data, _ := json.Marshal(c)

	fmt.Printf("%s", data)

	// Output:
	// {"age":2,"name":"Tiddles"}
}
